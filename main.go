package main

import (
	"bytes"
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type ssmWrapper struct {
	client *ssm.SSM
	cache  map[string]interface{}
}

func newSsmWrapper(region string) (*ssmWrapper, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	config := aws.NewConfig()
	if len(region) > 0 {
		config = aws.NewConfig().WithRegion(region)
	}

	return &ssmWrapper{ssm.New(sess, config), make(map[string]interface{})}, nil
}

func (w *ssmWrapper) Parameter(key string) (string, error) {
	if v, ok := w.cache[key]; ok {
		return v.(string), nil
	}

	input := &ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(true),
	}

	output, err := w.client.GetParameter(input)
	if err != nil {
		return "", err
	}

	w.cache[key] = *output.Parameter.Value

	return *output.Parameter.Value, nil
}

func (w *ssmWrapper) ParametersByPath(path string) (map[string]string, error) {
	if v, ok := w.cache[path]; ok {
		return v.(map[string]string), nil
	}

	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
	}

	output, err := w.client.GetParametersByPath(input)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)

	for _, p := range output.Parameters {
		m[*p.Name] = *p.Value
	}

	w.cache[path] = m

	return m, nil
}

func readTemplateFromStdin(fallback string) (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fallback, nil
	}

	t, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(t)), nil
}

func execTemplate(t *template.Template, w *ssmWrapper) (io.Reader, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, w); err != nil {
		return nil, err
	}
	return &buf, nil
}

func main() {
	var (
		region           = flag.String("region", "", "AWS region")
		tmplTextFallback = ""
	)
	flag.Parse()

	if args := flag.Args(); len(args) > 0 {
		tmplTextFallback = strings.TrimSpace(args[0])
	}

	tmplText, err := readTemplateFromStdin(tmplTextFallback)
	if err != nil {
		log.Fatalf("could not read stdin: %v\n", err)
	}

	if len(tmplText) == 0 {
		log.Fatalln("template not provided")
	}

	t, err := template.New("").Parse(tmplText)
	if err != nil {
		log.Fatalf("could not parse template %v\n", err)
	}

	w, err := newSsmWrapper(*region)
	if err != nil {
		log.Fatalf("could not create ssm client: %v", err)
	}

	r, err := execTemplate(t, w)
	if err != nil {
		log.Fatalf("could not process template: %v\n", err)
	}

	io.Copy(os.Stdout, r)
}
