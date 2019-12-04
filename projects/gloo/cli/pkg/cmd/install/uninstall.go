package install

import (
	"fmt"
	"io"
	"os"

	"github.com/solo-io/gloo/pkg/cliutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/solo-io/gloo/pkg/cliutil/install"
	"github.com/solo-io/gloo/projects/gloo/cli/pkg/cmd/options"
	"github.com/solo-io/gloo/projects/gloo/cli/pkg/constants"
)

func UninstallGloo(opts *options.Options, cli install.KubeCli) error {
	uninstaller := NewUninstaller(DefaultHelmClient(), cli)
	if err := uninstaller.Uninstall(opts); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Uninstall failed. Detailed logs available at %s.\n", cliutil.GetLogsPath())
		return err
	}
	return nil
}

type Uninstaller interface {
	Uninstall(cliArgs *options.Options) error
}

type uninstaller struct {
	helmClient HelmClient
	kubeCli    install.KubeCli
	output     io.Writer
}

func NewUninstaller(helmClient HelmClient, kubeCli install.KubeCli) Uninstaller {
	return NewUninstallerWithOutput(helmClient, kubeCli, os.Stdout)
}

// visible for testing
func NewUninstallerWithOutput(helmClient HelmClient, kubeCli install.KubeCli, output io.Writer) Uninstaller {
	return &uninstaller{
		helmClient: helmClient,
		kubeCli:    kubeCli,
		output:     output,
	}
}

func (u *uninstaller) Uninstall(cliArgs *options.Options) error {
	namespace := cliArgs.Uninstall.Namespace

	if releaseExists, err := ReleaseExists(u.helmClient, namespace, constants.GlooReleaseName); err != nil {
		return err
	} else if !releaseExists {
		fmt.Fprintf(u.output, "No Gloo installation found in namespace %s\n", namespace)
		return nil
	}

	uninstallAction, err := u.helmClient.NewUninstall(namespace)
	if err != nil {
		return err
	}

	var crdNames []string

	// need to run this first, as it depends on the release still being present
	if cliArgs.Uninstall.DeleteCrds {
		crdNames, err = u.findCrdNamesForRelease(namespace, constants.GlooReleaseName)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(u.output, "Removing Gloo system components from namespace %s...\n", namespace)
	if _, err = uninstallAction.Run(constants.GlooReleaseName); err != nil {
		return err
	}

	u.uninstallKnativeIfNecessary()

	if cliArgs.Uninstall.DeleteCrds {
		err := u.deleteGlooCrds(crdNames)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *uninstaller) findCrdNamesForRelease(namespace, releaseName string) (crdNames []string, err error) {
	lister, err := u.helmClient.ReleaseList(namespace)
	if err != nil {
		return nil, err
	}
	releases, err := lister.Run()
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, NoReleaseForCRDs
	} else if len(releases) > 1 {
		return nil, MultipleReleasesForCRDs
	}

	rel := releases[0]
	for _, crd := range rel.Chart.CRDs() {
		resource, err := makeUnstructured(string(crd.Data))
		if err != nil {
			return nil, err
		}

		crdNames = append(crdNames, resource.GetName())
	}

	return crdNames, nil
}

// expects the Helm release to still be present
func (u *uninstaller) deleteGlooCrds(crdNames []string) error {
	if len(crdNames) == 0 {
		return nil
	}

	fmt.Fprintf(u.output, "Removing Gloo CRDs...\n")
	args := []string{"delete", "crd"}
	for _, crdName := range crdNames {
		args = append(args, crdName)
	}
	if err := u.kubeCli.Kubectl(nil, args...); err != nil {
		fmt.Fprintf(u.output, "Unable to delete Gloo CRDs. Continuing...\n")
	}

	return nil
}

func makeUnstructured(manifest string) (*unstructured.Unstructured, error) {
	jsn, err := yaml.YAMLToJSON([]byte(manifest))
	if err != nil {
		return nil, err
	}
	runtimeObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsn)
	if err != nil {
		return nil, err
	}
	return runtimeObj.(*unstructured.Unstructured), nil
}

func (u *uninstaller) uninstallKnativeIfNecessary() {
	_, installOpts, err := checkKnativeInstallation()
	if err != nil {
		fmt.Fprintf(u.output, "Finding knative installation\n")
		return
	}
	if installOpts != nil {
		fmt.Fprintf(u.output, "Removing knative components installed by Gloo %#v...\n", installOpts)
		manifests, err := RenderKnativeManifests(*installOpts)
		if err != nil {
			fmt.Fprintf(u.output, "Could not determine which knative components to remove. Continuing...\n")
			return
		}
		if err := install.KubectlDelete([]byte(manifests), "--ignore-not-found"); err != nil {
			fmt.Fprintf(u.output, "Unable to delete knative. Continuing...\n")
		}
	}
}
