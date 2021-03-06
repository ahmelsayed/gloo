package upstreamssl_test

import (
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/plugins"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/gloo/projects/gloo/pkg/plugins/upstreamssl"
)

var _ = Describe("Plugin", func() {

	var (
		params   plugins.Params
		plugin   *Plugin
		upstream *v1.Upstream
		tlsConf  *v1.TlsSecret
		out      *envoyapi.Cluster
	)
	BeforeEach(func() {
		out = new(envoyapi.Cluster)

		tlsConf = &v1.TlsSecret{}
		params = plugins.Params{
			Snapshot: &v1.ApiSnapshot{
				Secrets: v1.SecretList{{
					Metadata: core.Metadata{
						Name:      "name",
						Namespace: "namespace",
					},
					Kind: &v1.Secret_Tls{
						Tls: tlsConf,
					},
				}},
			},
		}
		ref := params.Snapshot.Secrets[0].Metadata.Ref()

		upstream = &v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				SslConfig: &v1.UpstreamSslConfig{
					SslSecrets: &v1.UpstreamSslConfig_SecretRef{
						SecretRef: &ref,
					},
				},
			},
		}
		plugin = NewPlugin()
	})

	It("should process an upstream with tls config", func() {

		err := plugin.ProcessUpstream(params, upstream, out)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.TlsContext).ToNot(BeNil())
	})

	It("should process an upstream with tls config", func() {

		tlsConf.PrivateKey = "private"
		tlsConf.CertChain = "certchain"

		err := plugin.ProcessUpstream(params, upstream, out)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.TlsContext).ToNot(BeNil())
		Expect(out.TlsContext.CommonTlsContext.TlsCertificates[0].PrivateKey.GetInlineString()).To(Equal("private"))
		Expect(out.TlsContext.CommonTlsContext.TlsCertificates[0].CertificateChain.GetInlineString()).To(Equal("certchain"))
	})

	It("should process an upstream with rootca", func() {
		tlsConf.RootCa = "rootca"

		err := plugin.ProcessUpstream(params, upstream, out)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.TlsContext).ToNot(BeNil())
		Expect(out.TlsContext.CommonTlsContext.GetValidationContext().TrustedCa.GetInlineString()).To(Equal("rootca"))
	})

	Context("failure", func() {

		It("should fail with only private key", func() {

			tlsConf.PrivateKey = "private"

			err := plugin.ProcessUpstream(params, upstream, out)
			Expect(err).To(HaveOccurred())
		})
		It("should fail with only cert chain", func() {

			tlsConf.CertChain = "certchain"

			err := plugin.ProcessUpstream(params, upstream, out)
			Expect(err).To(HaveOccurred())
		})
	})
})
