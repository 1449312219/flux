package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"

	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/history"
	"github.com/weaveworks/flux/update"
)

const (
	defaultReleaseTemplate = `Release {{with .Release.Cause}}({{if .User}}{{.User}}{{else}}unknown{{end}}{{if .Message}}: {{.Message}}{{end}}){{end}} {{trim (print .Release.Spec.ImageSpec) "<>"}} to {{with .Release.Spec.ServiceSpecs}}{{range $index, $spec := .}}{{if not (eq $index 0)}}, {{if last $index $.Release.Spec.ServiceSpecs}}and {{end}}{{end}}{{trim (print .) "<>"}}{{end}}{{end}}. {{with .Error}}{{.}}. failed{{else}}done{{end}}`
)

var (
	httpClient = &http.Client{Timeout: 5 * time.Second}
)

func slackNotifyRelease(config flux.NotifierConfig, release *history.ReleaseEventMetadata, releaseError string) error {
	if release.Spec.Kind == update.ReleaseKindPlan {
		return nil
	}

	template := defaultReleaseTemplate
	if config.ReleaseTemplate != "" {
		template = config.ReleaseTemplate
	}

	errorMessage := ""
	if releaseError != "" {
		errorMessage = releaseError
	}
	text, err := instantiateTemplate("release", template, struct {
		Release *history.ReleaseEventMetadata
		Error   string
	}{
		Release: release,
		Error:   errorMessage,
	})
	if err != nil {
		return err
	}

	return notify(config, text)
}

func notify(config flux.NotifierConfig, text string) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(map[string]string{
		"username": config.Username,
		"text":     text,
	}); err != nil {
		return errors.Wrap(err, "encoding Slack POST request")
	}

	req, err := http.NewRequest("POST", config.HookURL, buf)
	if err != nil {
		return errors.Wrap(err, "constructing Slack HTTP request")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "executing HTTP POST to Slack")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		return fmt.Errorf("%s from Slack (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

func instantiateTemplate(tmplName, tmplStr string, args interface{}) (string, error) {
	tmpl, err := template.New(tmplName).Funcs(templateFuncs).Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, args); err != nil {
		return "", err
	}
	return buf.String(), nil
}
