package k8sadmisssion

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/solo-io/gloo/pkg/utils/skutils"

	validationapi "github.com/solo-io/gloo/projects/gloo/pkg/api/grpc/validation"

	"github.com/pkg/errors"
	gwv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	v2 "github.com/solo-io/gloo/projects/gateway/pkg/api/v2"
	"github.com/solo-io/gloo/projects/gateway/pkg/validation"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	ValidationPath    = "/validation"
	skipValidationKey = "gateway.solo.io/skip_validation"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

func skipValidationCheck(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}
	return annotations[skipValidationKey] == "true"
}

func NewGatewayValidatingWebhook(ctx context.Context, validator validation.Validator, watchNamespaces []string, port int, serverCertPath, serverKeyPath string) (*http.Server, error) {
	keyPair, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		return nil, errors.Wrapf(err, "loading x509 key pair")
	}

	handler := &gatewayValidationWebhook{
		ctx:             contextutils.WithLogger(ctx, "gateway-validation-webhook"),
		validator:       validator,
		watchNamespaces: watchNamespaces,
	}

	mux := http.NewServeMux()
	mux.Handle(ValidationPath, handler)

	return &http.Server{
		Addr:      fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{keyPair}},
		Handler:   mux,
	}, nil

}

type gatewayValidationWebhook struct {
	ctx             context.Context
	validator       validation.Validator
	watchNamespaces []string
}

func (wh *gatewayValidationWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := contextutils.LoggerFrom(wh.ctx)

	logger.Infow("received validation request")

	// Verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Errorf("contentType=%s, expecting application/json", contentType)
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
		defer r.Body.Close()
	}
	if len(body) == 0 {
		logger.Errorf("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	var (
		admissionResponse *v1beta1.AdmissionResponse
		review            v1beta1.AdmissionReview
	)
	if _, _, err := deserializer.Decode(body, nil, &review); err != nil {
		logger.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		admissionResponse = wh.validate(wh.ctx, &review)
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if review.Request != nil {
			admissionReview.Response.UID = review.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		logger.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}
	logger.Infof("Ready to write response ...")
	if _, err := w.Write(resp); err != nil {
		logger.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}

	logger.Infof("responded with review: %s", resp)
}
func (wh *gatewayValidationWebhook) validate(ctx context.Context, review *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	logger := contextutils.LoggerFrom(ctx)

	req := review.Request

	logger.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	gvk := schema.GroupVersionKind{
		Group:   req.Kind.Group,
		Version: req.Kind.Version,
		Kind:    req.Kind.Kind,
	}

	var validationErr error

	isDelete := req.Operation == v1beta1.Delete

	// ensure the request applies to a watched namespace, if watchNamespaces is set
	var validatingForNamespace bool
	if len(wh.watchNamespaces) > 0 {
		for _, ns := range wh.watchNamespaces {
			if ns == metav1.NamespaceAll || ns == req.Namespace {
				validatingForNamespace = true
				break
			}
		}
	} else {
		validatingForNamespace = true
	}

	// if it's not our namespace, do not validate
	if !validatingForNamespace {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	ref := core.ResourceRef{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	var proxyReports validation.ProxyReports
	switch gvk {
	case v2.GatewayGVK:
		if isDelete {
			// we don't validate gateway deletion
			break
		}
		proxyReports, validationErr = wh.validateGateway(ctx, req.Object.Raw)
	case gwv1.VirtualServiceGVK:
		if isDelete {
			validationErr = wh.validator.ValidateDeleteVirtualService(ctx, ref)
		} else {
			proxyReports, validationErr = wh.validateVirtualService(ctx, req.Object.Raw)
		}
	case gwv1.RouteTableGVK:
		if isDelete {
			validationErr = wh.validator.ValidateDeleteRouteTable(ctx, ref)
		} else {
			proxyReports, validationErr = wh.validateRouteTable(ctx, req.Object.Raw)
		}
	}

	if validationErr == nil {
		logger.Debug("Succeeded")

		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	logger.Errorf("Validation failed: %v", validationErr)

	details := &metav1.StatusDetails{
		Name:  req.Name,
		Group: gvk.Group,
		Kind:  gvk.Kind,
	}

	for _, proxyReport := range proxyReports {
		for _, listenerReport := range proxyReport.ListenerReports {
			for _, err := range listenerReport.Errors {
				details.Causes = append(details.Causes, metav1.StatusCause{
					Message: fmt.Sprintf("Listener Error %v: %v", err.Type.String(), err.Reason),
				})
			}
			switch listener := listenerReport.ListenerTypeReport.(type) {
			case *validationapi.ListenerReport_HttpListenerReport:
				for _, err := range listener.HttpListenerReport.Errors {
					details.Causes = append(details.Causes, metav1.StatusCause{
						Message: fmt.Sprintf("HTTPListener Error %v: %v", err.Type.String(), err.Reason),
					})
				}
				for _, vh := range listener.HttpListenerReport.VirtualHostReports {
					for _, err := range vh.Errors {
						details.Causes = append(details.Causes, metav1.StatusCause{
							Message: fmt.Sprintf("VirtualHost Error %v: %v", err.Type.String(), err.Reason),
						})
					}
					for _, r := range vh.RouteReports {
						for _, err := range r.Errors {
							details.Causes = append(details.Causes, metav1.StatusCause{
								Message: fmt.Sprintf("Route Error %v: %v", err.Type.String(), err.Reason),
							})
						}
					}
				}
			case *validationapi.ListenerReport_TcpListenerReport:
				for _, err := range listener.TcpListenerReport.Errors {
					details.Causes = append(details.Causes, metav1.StatusCause{
						Message: fmt.Sprintf("TCPListener Error %v: %v", err.Type.String(), err.Reason),
					})
				}
				for _, host := range listener.TcpListenerReport.TcpHostReports {
					for _, err := range host.Errors {
						details.Causes = append(details.Causes, metav1.StatusCause{
							Message: fmt.Sprintf("TcpHost Error %v: %v", err.Type.String(), err.Reason),
						})
					}
				}
			}
		}
	}

	if len(proxyReports) > 0 {
		// no need to duplicate the error message
		validationErr = errors.Errorf("resource incompatible with current Gloo snapshot")
	}

	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: validationErr.Error(),
			Details: details,
		},
	}

}

func (wh *gatewayValidationWebhook) validateGateway(ctx context.Context, rawJson []byte) (validation.ProxyReports, error) {
	var gw v2.Gateway
	if err := skutils.UnmarshalResource(rawJson, &gw); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal raw object")
	}
	if skipValidationCheck(gw.Metadata.Annotations) {
		return nil, nil
	}
	if proxyReports, err := wh.validator.ValidateGateway(ctx, &gw); err != nil {
		return proxyReports, errors.Wrapf(err, "Validating %T failed", gw)
	}
	return nil, nil
}

func (wh *gatewayValidationWebhook) validateVirtualService(ctx context.Context, rawJson []byte) (validation.ProxyReports, error) {
	var vs gwv1.VirtualService
	if err := skutils.UnmarshalResource(rawJson, &vs); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal raw object")
	}
	if skipValidationCheck(vs.Metadata.Annotations) {
		return nil, nil
	}
	if proxyReports, err := wh.validator.ValidateVirtualService(ctx, &vs); err != nil {
		return proxyReports, errors.Wrapf(err, "Validating %T failed", vs)
	}
	return nil, nil
}

func (wh *gatewayValidationWebhook) validateRouteTable(ctx context.Context, rawJson []byte) (validation.ProxyReports, error) {
	var rt gwv1.RouteTable
	if err := skutils.UnmarshalResource(rawJson, &rt); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal raw object")
	}
	if skipValidationCheck(rt.Metadata.Annotations) {
		return nil, nil
	}
	if proxyReports, err := wh.validator.ValidateRouteTable(ctx, &rt); err != nil {
		return proxyReports, errors.Wrapf(err, "Validating %T failed", rt)
	}
	return nil, nil
}
