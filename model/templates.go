package model

import (
	"bytes"
	"go/format"
	"io"
	"text/template"
)

// Generate writes a Go file with a client for using the parsed service.
func Generate(w io.Writer) error {
	t := template.New("root").Funcs(template.FuncMap{
		"godoc":      godoc,
		"exportable": exportable,
	})
	template.Must(common(t))
	template.Must(jsonClient(t))
	template.Must(queryClient(t))
	template.Must(ec2Client(t))
	template.Must(restCommon(t))
	template.Must(restXMLClient(t))
	template.Must(restJSONClient(t))

	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, service.Metadata.Protocol, service); err != nil {
		return err
	}

	b, err := format.Source(out.Bytes())
	if err != nil {
		_, _ = w.Write(out.Bytes())
		return err
	}

	_, err = w.Write(b)
	return err
}

func common(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "header" }}
// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

// Package {{ .PackageName }} provides a client for {{ .FullName }}.
package {{ .PackageName }}

import (
  "net/http"
  "time"

  "github.com/stripe/aws-go/aws"
  "github.com/stripe/aws-go/gen/endpoints"
)

{{ end }}

{{ define "footer" }}
// avoid errors if the packages aren't referenced
var _ time.Time
{{ end }}

`)
}

func jsonClient(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "json" }}
{{ template "header" $ }}

// {{ .Name }} is a client for {{ .FullName }}.
type {{ .Name }} struct {
  client *aws.JSONClient
}

// New returns a new {{ .Name }} client.
func New(creds aws.CredentialsProvider, region string, client *http.Client) *{{ .Name }} {
  if client == nil {
     client = http.DefaultClient
  }

  endpoint, service, region := endpoints.Lookup("{{ .Metadata.EndpointPrefix }}", region)

  return &{{ .Name }}{
    client: &aws.JSONClient{
      Context: aws.Context{
        Credentials: creds,
        Service: service,
        Region: region,
      },      Client: client,
      Endpoint: endpoint,
      JSONVersion: "{{ .Metadata.JSONVersion }}",
      TargetPrefix: "{{ .Metadata.TargetPrefix }}",
    },
  }
}

{{ range $name, $op := .Operations }}

{{ godoc $name $op.Documentation }} func (c *{{ $.Name }}) {{ exportable $name }}({{ if $op.Input }}req {{ $op.Input.Type }}{{ end }}) ({{ if $op.Output }}resp {{ $op.Output.Type }},{{ end }} err error) {
  {{ if $op.Output }}resp = {{ $op.Output.Literal }}{{ else }}// NRE{{ end }}
  err = c.client.Do("{{ $name }}", "{{ $op.HTTP.Method }}", "{{ $op.HTTP.RequestURI }}", {{ if $op.Input }} req {{ else }} nil {{ end }}, {{ if $op.Output }} resp {{ else }} nil {{ end }})
  return
}

{{ end }}

{{ range $name, $s := .Shapes }}
{{ if eq $s.ShapeType "structure" }}
{{ if not $s.Exception }}

// {{ exportable $name }} is undocumented.
type {{ exportable $name }} struct {
{{ range $name, $m := $s.Members }}
{{ exportable $name }} {{ $m.Type }} {{ $m.JSONTag }}  {{ end }}
}

{{ end }}
{{ else if $s.Enum }}
// Possible values for {{ $.Name }}.
const (
{{ range $name, $value := $s.Enums }}
{{ $name }} = {{ $value }}{{ end }}
)
{{ end }}
{{ end }}

{{ template "footer" }}
{{ end }}

`)
}

func queryClient(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "query" }}
{{ template "header" $ }}

// {{ .Name }} is a client for {{ .FullName }}.
type {{ .Name }} struct {
  client *aws.QueryClient
}

// New returns a new {{ .Name }} client.
func New(creds aws.CredentialsProvider, region string, client *http.Client) *{{ .Name }} {
  if client == nil {
     client = http.DefaultClient
  }

  endpoint, service, region := endpoints.Lookup("{{ .Metadata.EndpointPrefix }}", region)

  return &{{ .Name }}{
    client: &aws.QueryClient{
      Context: aws.Context{
        Credentials: creds,
        Service: service,
        Region: region,
      },
      Client: client,
      Endpoint: endpoint,
      APIVersion: "{{ .Metadata.APIVersion }}",
    },
  }
}

{{ range $name, $op := .Operations }}

{{ godoc $name $op.Documentation }} func (c *{{ $.Name }}) {{ exportable $name }}({{ if $op.InputRef }}req {{ $op.InputRef.WrappedType }}{{ end }}) ({{ if $op.OutputRef }}resp {{ $op.OutputRef.WrappedType }},{{ end }} err error) {
  {{ if $op.Output }}resp = {{ $op.OutputRef.WrappedLiteral }}{{ else }}// NRE{{ end }}
  err = c.client.Do("{{ $name }}", "{{ $op.HTTP.Method }}", "{{ $op.HTTP.RequestURI }}", {{ if $op.Input }} req {{ else }} nil {{ end }}, {{ if $op.Output }} resp {{ else }} nil {{ end }})
  return
}

{{ end }}

{{ range $name, $s := .Shapes }}
{{ if eq $s.ShapeType "structure" }}
{{ if not $s.Exception }}

// {{ exportable $name }} is undocumented.
type {{ exportable $name }} struct {
{{ range $name, $m := $s.Members }}
{{ exportable $name }} {{ $m.Type }} {{ $m.QueryTag $s.ResultWrapper }}  {{ end }}
}

{{ end }}
{{ else if $s.Enum }}
// Possible values for {{ $.Name }}.
const (
{{ range $name, $value := $s.Enums }}
{{ $name }} = {{ $value }}{{ end }}
)
{{ end }}
{{ end }}

{{ range $wname, $s := .Wrappers }}

// {{ exportable $wname }} is a wrapper for {{ $s.Name }}.
type {{ exportable $wname }} struct {
{{ range $name, $m := $s.Members }}
{{ exportable $name }} {{ $m.Type }} {{ $m.QueryTag $wname }}  {{ end }}
}

{{ end }}

{{ template "footer" }}
{{ end }}

`)
}

func ec2Client(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "ec2" }}
{{ template "header" $ }}

// {{ .Name }} is a client for {{ .FullName }}.
type {{ .Name }} struct {
  client *aws.EC2Client
}

// New returns a new {{ .Name }} client.
func New(creds aws.CredentialsProvider, region string, client *http.Client) *{{ .Name }} {
  if client == nil {
     client = http.DefaultClient
  }

  endpoint, service, region := endpoints.Lookup("{{ .Metadata.EndpointPrefix }}", region)

  return &{{ .Name }}{
    client: &aws.EC2Client{
      Context: aws.Context{
        Credentials: creds,
        Service: service,
        Region: region,
      },
      Client: client,
      Endpoint: endpoint,
      APIVersion: "{{ .Metadata.APIVersion }}",
    },
  }
}

{{ range $name, $op := .Operations }}

{{ godoc $name $op.Documentation }} func (c *{{ $.Name }}) {{ exportable $name }}({{ if $op.InputRef }}req {{ $op.InputRef.WrappedType }}{{ end }}) ({{ if $op.OutputRef }}resp {{ $op.OutputRef.WrappedType }},{{ end }} err error) {
  {{ if $op.Output }}resp = {{ $op.OutputRef.WrappedLiteral }}{{ else }}// NRE{{ end }}
  err = c.client.Do("{{ $name }}", "{{ $op.HTTP.Method }}", "{{ $op.HTTP.RequestURI }}", {{ if $op.Input }} req {{ else }} nil {{ end }}, {{ if $op.Output }} resp {{ else }} nil {{ end }})
  return
}

{{ end }}

{{ range $name, $s := .Shapes }}
{{ if eq $s.ShapeType "structure" }}
{{ if not $s.Exception }}

// {{ exportable $name }} is undocumented.
type {{ exportable $name }} struct {
{{ range $name, $m := $s.Members }}
{{ exportable $name }} {{ $m.Type }} {{ $m.EC2Tag }}  {{ end }}
}

{{ end }}
{{ else if $s.Enum }}
// Possible values for {{ $.Name }}.
const (
{{ range $name, $value := $s.Enums }}
{{ $name }} = {{ $value }}{{ end }}
)
{{ end }}
{{ end }}

{{ range $wname, $s := .Wrappers }}

// {{ exportable $wname }} is a wrapper for {{ $s.Name }}.
type {{ exportable $wname }} struct {
{{ range $name, $m := $s.Members }}
{{ exportable $name }} {{ $m.Type }} {{ $m.EC2Tag }}  {{ end }}
}

{{ end }}

{{ template "footer" }}
{{ end }}

`)
}

func restCommon(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "rest-uri" }}
  {{ if .Input }}
  {{ range $name, $m := .Input.Members }}
  {{ if eq $m.Location "uri" }}

  if req.{{ exportable $name }} != nil {
    uri = strings.Replace(uri, "{"+"{{ $m.LocationName }}"+"}", aws.EscapePath(*req.{{ exportable $name }}), -1)
    uri = strings.Replace(uri, "{"+"{{ $m.LocationName }}+"+"}", aws.EscapePath(*req.{{ exportable $name }}), -1)
  }

  {{ end }}
  {{ end }}
  {{ end }}

{{ end }}


{{ define "rest-querystring" }}
  q := url.Values{}

  {{ if .Input }}
  {{ range $name, $m := .Input.Members }}
  {{ if eq $m.Location "querystring" }}


  {{ if eq $m.Shape.ShapeType "string" }}

  if req.{{ exportable $name }} != nil {
    q.Set("{{ $m.LocationName }}", *req.{{ exportable $name }})
  }

  {{ else if eq $m.Shape.ShapeType "timestamp" }}

  if req.{{ exportable $name }} != (time.Time{}) {
    q.Set("{{ $m.LocationName }}", req.{{ exportable $name }}.Format(time.RFC822))
  }

  {{ else if eq $m.Shape.ShapeType "integer" }}

  if req.{{ exportable $name }} != nil {
    q.Set("{{ $m.LocationName }}", strconv.Itoa(*req.{{ exportable $name }}))
  }

  {{ else }}

  if req.{{ exportable $name }} != nil {
    q.Set("{{ $m.LocationName }}", fmt.Sprintf("%v", req.{{ exportable $name }}))
  }

  {{ end }}

  {{ end }}
  {{ end }}
  {{ end }}

  if len(q) > 0 {
    uri += "?" + q.Encode()
  }
{{ end }}

{{ define "rest-reqheaders" }}
  {{ if .Input }}
  {{ range $name, $m := .Input.Members }}
  {{ if eq $m.Location "header" }}

 {{ if eq $m.Shape.ShapeType "string" }}

  if req.{{ exportable $name }} != nil {
    httpReq.Header.Set("{{ $m.LocationName }}", *req.{{ exportable $name }})
  }

  {{ else if eq $m.Shape.ShapeType "timestamp" }}

  if req.{{ exportable $name }} != (time.Time{}) {
    httpReq.Header.Set("{{ $m.LocationName }}", req.{{ exportable $name }}.Format(time.RFC822))
  }

  {{ else if eq $m.Shape.ShapeType "integer" }}

  {{ if eq $m.LocationName "Content-Length" }}

  if req.{{ exportable $name }} != nil {
    httpReq.ContentLength = *req.{{ exportable $name }}
  }

  {{ else }}

  if req.{{ exportable $name }} != nil {
    httpReq.Header.Set("{{ $m.LocationName }}", strconv.Itoa(*req.{{ exportable $name }}))
  }

  {{ end }}

  {{ else }}

  if req.{{ exportable $name }} != nil {
    httpReq.Header.Set("{{ $m.LocationName }}", fmt.Sprintf("%v", req.{{ exportable $name }}))
  }

  {{ end }}

  {{ else if eq $m.Location "headers" }}

  for name, value := range req.{{ exportable $name }} {
	httpReq.Header.Set(name, value)
  }

  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ define "rest-respheaders" }}
 {{ range $name, $m := .Output.Members }}
    {{ if ne $name "Body" }}
      {{ if eq $m.Location "header" }}
        if s := httpResp.Header.Get("{{ $m.LocationName }}"); s != "" {
         {{ if eq $m.Shape.ShapeType "string" }}
          resp.{{ exportable $name }} = &s
         {{ else if eq $m.Shape.ShapeType "timestamp" }}
           var t time.Time
           t, err = time.Parse(time.RFC1123, s)
           if err != nil {
             return
           }
           resp.{{ exportable $name }} = t
         {{ else if eq $m.Shape.ShapeType "integer" }}
           {{ if eq $m.Shape.Name "ContentLength" }}
           var n int64
           n, err = strconv.ParseInt(s, 10, 64)
           {{ else }}
           var n int
           n, err = strconv.Atoi(s)
           {{ end }}
           if  err != nil {
             return
           }
           resp.{{ exportable $name }} = &n
         {{ else if eq $m.Shape.ShapeType "boolean" }}
           var v bool
           v, err = strconv.ParseBool(s)
           if  err != nil {
             return
            }
           resp.{{ exportable $name }} = &v
         {{ else }}
         // TODO: add support for {{ $m.Shape.ShapeType }} headers
         {{ end }}
        }
      {{ else if eq $m.Location "headers" }}
      resp.{{ exportable $name }} = {{ $m.Shape.Type }}{}
      for name := range httpResp.Header {
        if strings.HasPrefix(name, "X-Amz-Meta-") {
          resp.{{ exportable $name }}[name] = httpResp.Header.Get(name)
        }
      }
      {{ else if eq $m.Location "statusCode" }}
        resp.{{ exportable $name }} = aws.Integer(httpResp.StatusCode)
      {{ else if ne $m.Location "" }}
      // TODO: add support for extracting output members from {{ $m.Location }} to support {{ exportable $name }}
      {{ end }}

    {{ end }}
  {{ end }}
{{ end }}
`)
}

func restXMLClient(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "rest-xml" }}
{{ template "header" $ }}

import (
  "bytes"
  "encoding/xml"
  "fmt"
  "io"
  "io/ioutil"
  "net/url"
  "strconv"
  "strings"
)

// {{ .Name }} is a client for {{ .FullName }}.
type {{ .Name }} struct {
  client *aws.RestClient
}

// New returns a new {{ .Name }} client.
func New(creds aws.CredentialsProvider, region string, client *http.Client) *{{ .Name }} {
  if client == nil {
     client = http.DefaultClient
  }

  endpoint, service, region := endpoints.Lookup("{{ .Metadata.EndpointPrefix }}", region)

  return &{{ .Name }}{
    client: &aws.RestClient{
      Context: aws.Context{
        Credentials: creds,
        Service: service,
        Region: region,
      },
      Client: client,
      Endpoint: endpoint,
      APIVersion: "{{ .Metadata.APIVersion }}",
    },
  }
}

{{ range $name, $op := .Operations }}

{{ godoc $name $op.Documentation }} func (c *{{ $.Name }}) {{ exportable $name }}({{ if $op.Input }}req {{ $op.Input.Type }}{{ end }}) ({{ if $op.Output }}resp {{ $op.Output.Type }},{{ end }} err error) {
  {{ if $op.Output }}resp = {{ $op.Output.Literal }}{{ else }}// NRE{{ end }}

  var body io.Reader
  var contentType string
  {{ if $op.Input }}

  {{ if $op.Input.Payload }}
  {{ with $m := index $op.Input.Members $op.Input.Payload }}
  {{ if $m.Streaming }}
  body = req.{{ exportable $m.Name  }}
  {{ else }}
  contentType = "application/xml"
	{{ if ne $m.LocationName ""}}
  if req.{{ exportable $m.Name }} != nil {
	req.{{ exportable $m.Name }}.XMLName = xml.Name{
		Space: "{{ $m.XMLNamespace.URI }}",
		Local: "{{ $m.LocationName }}",
	}
  }
  {{ end }}
  b, err := xml.Marshal(req.{{ exportable $m.Name }})
  if err != nil {
    return
  }
  body = bytes.NewReader(b)
  {{ end }}
  {{ end }}
  {{ else if $op.InputRef.LocationName }}
  contentType = "application/xml"
  req.XMLName = xml.Name{
		Space: "{{ $op.InputRef.XMLNamespace.URI }}",
		Local: "{{ $op.InputRef.LocationName }}",
	}
  b, err := xml.Marshal(req)
  if err != nil {
    return
  }
  body = bytes.NewReader(b)
  {{ end }}
  {{ end }}


  uri := c.client.Endpoint + "{{ $op.HTTP.RequestURI }}"
  {{ template "rest-uri" $op }}


  {{ template "rest-querystring" $op }}

  httpReq, err := http.NewRequest("{{ $op.HTTP.Method }}", uri, body)
  if err != nil {
    return
  }

  if contentType != "" {
    httpReq.Header.Set("Content-Type", contentType)
  }

  {{ template "rest-reqheaders" $op }}

  httpResp, err := c.client.Do(httpReq)
  if err != nil {
    return
  }

  {{ if $op.Output }}
    {{ with $name := "Body" }}
    {{ with $m := index $op.Output.Members $name }}
    {{ if $m }}

      {{ if $m.Streaming }}
  resp.Body = httpResp.Body
      {{ else }}
  defer httpResp.Body.Close()
  if e := xml.NewDecoder(httpResp.Body).Decode(resp); e != nil && e != io.EOF {
    err = e
    return
  }
      {{ end }}


    {{ else }}
  defer httpResp.Body.Close()
    {{ end }}
    {{ end }}

  {{ template "rest-respheaders" $op }}
  {{ end }}
  {{ else }}
  defer httpResp.Body.Close()
  {{ end }}


  return
}

{{ end }}

{{ range $name, $s := .Shapes }}
{{ if eq $s.ShapeType "structure" }}
{{ if not $s.Exception }}

// {{ exportable $name }} is undocumented.
type {{ exportable $name }} struct {
  XMLName xml.Name
{{ range $name, $m := $s.Members }}
{{ exportable $name }} {{ $m.Type }} {{ $m.XMLTag $s.ResultWrapper }}  {{ end }}
}

func (v *{{ exportable $name }}) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return aws.MarshalXML(v, e, start)
}

{{ end }}
{{ else if $s.Enum }}
// Possible values for {{ $.Name }}.
const (
{{ range $name, $value := $s.Enums }}
{{ $name }} = {{ $value }}{{ end }}
)
{{ end }}
{{ end }}

{{ template "footer" }}
var _ bytes.Reader
var _ url.URL
var _ fmt.Stringer
var _ strings.Reader
var _ strconv.NumError
var _ = ioutil.Discard
{{ end }}

`)
}

func restJSONClient(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "rest-json" }}
{{ template "header" $ }}

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  "net/url"
  "strconv"
  "strings"
)

// {{ .Name }} is a client for {{ .FullName }}.
type {{ .Name }} struct {
  client *aws.RestClient
}

// New returns a new {{ .Name }} client.
func New(creds aws.CredentialsProvider, region string, client *http.Client) *{{ .Name }} {
  if client == nil {
     client = http.DefaultClient
  }

  endpoint, service, region := endpoints.Lookup("{{ .Metadata.EndpointPrefix }}", region)

  return &{{ .Name }}{
    client: &aws.RestClient{
      Context: aws.Context{
        Credentials: creds,
        Service: service,
        Region: region,
      },
      Client: client,
      Endpoint: endpoint,
      APIVersion: "{{ .Metadata.APIVersion }}",
    },
  }
}

{{ range $name, $op := .Operations }}

{{ godoc $name $op.Documentation }} func (c *{{ $.Name }}) {{ exportable $name }}({{ if $op.Input }}req {{ $op.Input.Type }}{{ end }}) ({{ if $op.Output }}resp {{ $op.Output.Type }},{{ end }} err error) {
  {{ if $op.Output }}resp = {{ $op.Output.Literal }}{{ else }}// NRE{{ end }}

  var body io.Reader
  var contentType string
  {{ if $op.Input }}
  {{ if $op.Input.Payload }}
  {{ with $m := index $op.Input.Members $op.Input.Payload }}
  {{ if $m.Streaming }}
  body = req.{{ exportable $m.Name  }}
  {{ else }}
  contentType = "application/json"
  b, err := json.Marshal(req.{{ exportable $m.Name  }})
  if err != nil {
    return
  }
  body = bytes.NewReader(b)
  {{ end }}
  {{ end }}
  {{ end }}
  {{ end }}


  uri := c.client.Endpoint + "{{ $op.HTTP.RequestURI }}"
  {{ template "rest-uri" $op }}

  {{ template "rest-querystring" $op }}

  httpReq, err := http.NewRequest("{{ $op.HTTP.Method }}", uri, body)
  if err != nil {
    return
  }

  if contentType != "" {
    httpReq.Header.Set("Content-Type", contentType)
  }

  {{ template "rest-reqheaders" $op }}

  httpResp, err := c.client.Do(httpReq)
  if err != nil {
    return
  }

  {{ if $op.Output }}
    {{ with $name := "Body" }}
    {{ with $m := index $op.Output.Members $name }}
    {{ if $m }}

      {{ if $m.Streaming }}
  resp.Body = httpResp.Body
      {{ else }}
  defer httpResp.Body.Close()

  if e := json.NewDecoder(httpResp.Body).Decode(resp); e != nil && e != io.EOF {
    err = e
    return
  }

      {{ end }}


    {{ else }}
  defer httpResp.Body.Close()
    {{ end }}
    {{ end }}

   {{ template "rest-respheaders" $op }}
  {{ end }}
  {{ else }}
  defer httpResp.Body.Close()
  {{ end }}


  return
}

{{ end }}

{{ range $name, $s := .Shapes }}
{{ if eq $s.ShapeType "structure" }}
{{ if not $s.Exception }}

// {{ exportable $name }} is undocumented.
type {{ exportable $name }} struct {
{{ range $name, $m := $s.Members }}
{{ exportable $name }} {{ $m.Type }} {{ $m.JSONTag }}  {{ end }}
}

{{ end }}
{{ else if $s.Enum }}
// Possible values for {{ $.Name }}.
const (
{{ range $name, $value := $s.Enums }}
{{ $name }} = {{ $value }}{{ end }}
)
{{ end }}
{{ end }}

{{ template "footer" }}
var _ bytes.Reader
var _ url.URL
var _ fmt.Stringer
var _ strings.Reader
var _ strconv.NumError
var _ = ioutil.Discard
var _ json.RawMessage
{{ end }}
`)
}
