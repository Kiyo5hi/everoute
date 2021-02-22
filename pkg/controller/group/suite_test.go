/*
Copyright 2021 The Lynx Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package group_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	agentv1alpha1 "github.com/smartxworks/lynx/pkg/apis/agent/v1alpha1"
	groupv1alpha1 "github.com/smartxworks/lynx/pkg/apis/group/v1alpha1"
	securityv1alpha1 "github.com/smartxworks/lynx/pkg/apis/security/v1alpha1"
	groupctrl "github.com/smartxworks/lynx/pkg/controller/group"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient          client.Client // You'll be using this client in your tests.
	testEnv            *envtest.Environment
	useExistingCluster bool
)

const (
	RunTestWithExistingCluster = "TESTING_WITH_EXISTING_CLUSTER"
)

func TestGroupController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "GroupController Suite", []Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	if os.Getenv(RunTestWithExistingCluster) == "true" {
		By("testing with existing cluster")
		useExistingCluster = true
	}

	/*
		First, the envtest cluster is configured to read CRDs from the CRD directory Kubebuilder scaffolds for you.
	*/
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		UseExistingCluster: &useExistingCluster,
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths:           []string{filepath.Join("..", "..", "..", "deploy", "crds")},
			CleanUpAfterUse: true,
		},
	}

	/*
		Then, we start the envtest cluster.
	*/
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	/*
		The autogenerated test code will add schema to the default client-go k8s scheme.
	*/
	err = securityv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = agentv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = groupv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	//+kubebuilder:scaffold:scheme

	/*
		One thing that this autogenerated file is missing, however, is a way to actually start your controller.
		The code above will set up a client for interacting with your custom Kind,
		but will not be able to test your controller behavior.
		If you want to test your custom controller logic, you’ll need to add some familiar-looking manager logic
		to your BeforeSuite() function, so you can register your custom controller to run on this test cluster.
		You may notice that the code below runs your controller with nearly identical logic to your CronJob project’s main.go!
		The only difference is that the manager is started in a separate goroutine so it does not block the cleanup of envtest
		when you’re done running your tests.
		Once you've added the code below, you can actually delete the k8sClient above, because you can get k8sClient from the manager
		(as shown below).
	*/

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sManager).ToNot(BeNil())

	err = (&groupctrl.GroupReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
