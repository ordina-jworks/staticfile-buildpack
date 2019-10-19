package integration_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("deploy a staticfile app with visitable directories", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(Fixtures("no_trailing_slash"))
	})

	It("redirects to https directory when path doesn't end with a slash", func() {
		PushAppAndConfirm(app)

		resp, err := GetResponse(app, "/something")
		Expect(err).To(BeNil())
		Expect(resp.TLS).NotTo(BeNil())
		Expect(resp.TLS.HandshakeComplete).To(BeTrue())
	})
})

func GetResponse(a *cutlass.App, path string) (*http.Response, error) {
	url, err := GetHttpsUrl(a, path)
	if err != nil {
		return new(http.Response), err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return client.Get(url)
}

func GetHttpsUrl(a *cutlass.App, path string) (string, error) {
	guid, err := a.AppGUID()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("cf", "curl", "/v2/apps/"+guid+"/summary")
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	host := gjson.Get(string(data), "routes.0.host").String()
	domain := gjson.Get(string(data), "routes.0.domain.name").String()
	return fmt.Sprintf("%s://%s.%s%s", "https", host, domain, path), nil
}
