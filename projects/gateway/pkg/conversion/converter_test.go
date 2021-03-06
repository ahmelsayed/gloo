package conversion_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	gatewayv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	gatewayv2 "github.com/solo-io/gloo/projects/gateway/pkg/api/v2"
	"github.com/solo-io/gloo/projects/gateway/pkg/conversion"
	"github.com/solo-io/gloo/projects/gateway/pkg/mocks/mock_conversion"
	"github.com/solo-io/gloo/projects/gateway/pkg/mocks/mock_v1"
	"github.com/solo-io/gloo/projects/gateway/pkg/mocks/mock_v2"
	"github.com/solo-io/gloo/projects/gloo/pkg/defaults"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	sk_errors "github.com/solo-io/solo-kit/pkg/errors"
)

var (
	resourceConverter conversion.ResourceConverter
	mockCtrl          *gomock.Controller
	v1GatewayClient   *mock_v1.MockGatewayClient
	v2GatewayClient   *mock_v2.MockGatewayClient
	gatewayConverter  *mock_conversion.MockGatewayConverter
	namespace         = "test-ns"
	testErr           = errors.Errorf("test-err")
	fooV1, barV1      *gatewayv1.Gateway
	fooV2, barV2      *gatewayv2.Gateway
)

var _ = Describe("ResourceConverter", func() {
	Describe("ConvertAll", func() {

		getV1Gateway := func(name string) *gatewayv1.Gateway {
			return &gatewayv1.Gateway{
				Metadata: core.Metadata{
					Namespace: namespace,
					Name:      name,
				},
			}
		}

		getV2Gateway := func(name string, resourceVersion string, annotations map[string]string) *gatewayv2.Gateway {
			return &gatewayv2.Gateway{
				Metadata: core.Metadata{
					Namespace:       namespace,
					Name:            name,
					ResourceVersion: resourceVersion,
					Annotations:     annotations,
				},
			}
		}

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			v1GatewayClient = mock_v1.NewMockGatewayClient(mockCtrl)
			v2GatewayClient = mock_v2.NewMockGatewayClient(mockCtrl)
			gatewayConverter = mock_conversion.NewMockGatewayConverter(mockCtrl)
			resourceConverter = conversion.NewResourceConverter(namespace, v1GatewayClient, v2GatewayClient, gatewayConverter)

			fooV1 = getV1Gateway("foo")
			barV1 = getV1Gateway("bar")
			fooV2 = getV2Gateway("foo", "", map[string]string{defaults.OriginKey: defaults.ConvertedValue})
			barV2 = getV2Gateway("bar", "", map[string]string{defaults.OriginKey: defaults.ConvertedValue})
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		Context("happy paths", func() {
			It("works when v2 resources don't exist already", func() {
				v1Gateways := []*gatewayv1.Gateway{fooV1, barV1}

				v1GatewayClient.EXPECT().
					List(namespace, clients.ListOpts{Ctx: context.TODO()}).
					Return(v1Gateways, nil)
				gatewayConverter.EXPECT().
					FromV1ToV2(fooV1).
					Return(fooV2)
				gatewayConverter.EXPECT().
					FromV1ToV2(barV1).
					Return(barV2)
				v2GatewayClient.EXPECT().
					Read(namespace, "foo", clients.ReadOpts{Ctx: context.TODO()}).
					Return(nil, sk_errors.NewNotExistErr(namespace, "foo", testErr))
				v2GatewayClient.EXPECT().
					Read(namespace, "bar", clients.ReadOpts{Ctx: context.TODO()}).
					Return(nil, sk_errors.NewNotExistErr(namespace, "bar", testErr))
				v2GatewayClient.EXPECT().
					Write(fooV2, clients.WriteOpts{Ctx: context.TODO(), OverwriteExisting: false}).
					Return(fooV2, nil)
				v2GatewayClient.EXPECT().
					Write(barV2, clients.WriteOpts{Ctx: context.TODO(), OverwriteExisting: false}).
					Return(barV2, nil)

				err := resourceConverter.ConvertAll(context.TODO())
				Expect(err).NotTo(HaveOccurred())
			})

			It("overwrites existing v2 resources which are marked as defaults", func() {
				existingFooV2 := getV2Gateway("foo", "10", map[string]string{defaults.OriginKey: defaults.DefaultValue})
				fooV2ToWrite := getV2Gateway("foo", "10", map[string]string{defaults.OriginKey: defaults.ConvertedValue})
				v1Gateways := []*gatewayv1.Gateway{fooV1, barV1}

				v1GatewayClient.EXPECT().
					List(namespace, clients.ListOpts{Ctx: context.TODO()}).
					Return(v1Gateways, nil)
				gatewayConverter.EXPECT().
					FromV1ToV2(fooV1).
					Return(fooV2)
				gatewayConverter.EXPECT().
					FromV1ToV2(barV1).
					Return(barV2)
				v2GatewayClient.EXPECT().
					Read(namespace, "foo", clients.ReadOpts{Ctx: context.TODO()}).
					Return(existingFooV2, nil)
				v2GatewayClient.EXPECT().
					Read(namespace, "bar", clients.ReadOpts{Ctx: context.TODO()}).
					Return(nil, sk_errors.NewNotExistErr(namespace, "bar", testErr))
				v2GatewayClient.EXPECT().
					Write(fooV2ToWrite, clients.WriteOpts{Ctx: context.TODO(), OverwriteExisting: true}).
					Return(fooV2ToWrite, nil)
				v2GatewayClient.EXPECT().
					Write(barV2, clients.WriteOpts{Ctx: context.TODO(), OverwriteExisting: false}).
					Return(barV2, nil)

				err := resourceConverter.ConvertAll(context.TODO())
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not overwrite existing resources which are marked as converted", func() {
				existingFooV2 := getV2Gateway("foo", "", map[string]string{defaults.OriginKey: defaults.ConvertedValue})
				v1Gateways := []*gatewayv1.Gateway{fooV1, barV1}

				v1GatewayClient.EXPECT().
					List(namespace, clients.ListOpts{Ctx: context.TODO()}).
					Return(v1Gateways, nil)
				gatewayConverter.EXPECT().
					FromV1ToV2(barV1).
					Return(barV2)
				gatewayConverter.EXPECT().
					FromV1ToV2(fooV1).
					Return(fooV2)
				v2GatewayClient.EXPECT().
					Read(namespace, "foo", clients.ReadOpts{Ctx: context.TODO()}).
					Return(existingFooV2, nil)
				v2GatewayClient.EXPECT().
					Read(namespace, "bar", clients.ReadOpts{Ctx: context.TODO()}).
					Return(nil, sk_errors.NewNotExistErr(namespace, "bar", testErr))
				v2GatewayClient.EXPECT().
					Write(barV2, clients.WriteOpts{Ctx: context.TODO(), OverwriteExisting: false}).
					Return(barV2, nil)

				err := resourceConverter.ConvertAll(context.TODO())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("sad paths", func() {
			It("errors if v1 gateway client errors on list", func() {
				v1GatewayClient.EXPECT().
					List(namespace, clients.ListOpts{Ctx: context.TODO()}).
					Return(nil, testErr)

				err := resourceConverter.ConvertAll(context.TODO())
				Expect(err).To(HaveOccurred())
				expectedErr := conversion.FailedToListGatewayResourcesError(err, "gatewayv1", namespace)
				Expect(expectedErr.Error()).To(ContainSubstring(err.Error()))
			})

			It("errors if v2 gateway client errors on read", func() {
				v1Gateways := []*gatewayv1.Gateway{fooV1}

				v1GatewayClient.EXPECT().
					List(namespace, clients.ListOpts{Ctx: context.TODO()}).
					Return(v1Gateways, nil)
				gatewayConverter.EXPECT().
					FromV1ToV2(fooV1).
					Return(fooV2)
				v2GatewayClient.EXPECT().
					Read(namespace, "foo", clients.ReadOpts{Ctx: context.TODO()}).
					Return(nil, testErr)

				err := resourceConverter.ConvertAll(context.TODO())
				Expect(err).To(HaveOccurred())
				expectedErr := conversion.FailedToReadExistingGatewayError(err, "v2", namespace, "foo")
				Expect(expectedErr.Error()).To(ContainSubstring(err.Error()))
			})

			It("errors if v2 gateway client errors on write", func() {
				v1Gateways := []*gatewayv1.Gateway{fooV1}

				v1GatewayClient.EXPECT().
					List(namespace, clients.ListOpts{Ctx: context.TODO()}).
					Return(v1Gateways, nil)
				gatewayConverter.EXPECT().
					FromV1ToV2(fooV1).
					Return(fooV2)
				v2GatewayClient.EXPECT().
					Read(namespace, "foo", clients.ReadOpts{Ctx: context.TODO()}).
					Return(nil, sk_errors.NewNotExistErr(namespace, "foo", testErr))
				v2GatewayClient.EXPECT().
					Write(fooV2, clients.WriteOpts{Ctx: context.TODO(), OverwriteExisting: false}).
					Return(nil, testErr)

				err := resourceConverter.ConvertAll(context.TODO())
				Expect(err).To(HaveOccurred())
				expectedErr := conversion.FailedToWriteGatewayError(err, "v2", namespace, "foo")
				Expect(expectedErr.Error()).To(ContainSubstring(err.Error()))
			})
		})
	})
})
