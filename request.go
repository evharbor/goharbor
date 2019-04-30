package goharbor

import (
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
	"github.com/levigross/grequests"
)

// RequestStruct 请求结构体
type RequestStruct struct {
	configs ConfigStruct
}

func getRequestURI(userURL string) (string, error) {
	u, err := url.Parse(userURL)
	if err != nil {
		return "", err
	}
	// fullpath := u.RequestURI()
	fullPath := u.Path
	query, err := url.QueryUnescape(u.RawQuery)
	if err != nil {
		return "", err
	}
	if query != "" {
		fullPath = fullPath + "?" + query
	}

	return fullPath, nil
}

// buildURLParams returns a URL with all of the params
// Note: This function will override current URL params if they contradict what is provided in the map
// That is what the "magic" is on the last line
func buildURLParams(userURL string, params map[string]string) (string, error) {
	parsedURL, err := url.Parse(userURL)

	if err != nil {
		return "", err
	}

	parsedQuery, err := url.ParseQuery(parsedURL.RawQuery)

	if err != nil {
		return "", nil
	}

	for key, value := range params {
		parsedQuery.Set(key, value)
	}

	return addQueryParams(parsedURL, parsedQuery), nil
}

func buildURLStruct(userURL string, URLStruct interface{}) (string, error) {
	parsedURL, err := url.Parse(userURL)

	if err != nil {
		return "", err
	}

	parsedQuery, err := url.ParseQuery(parsedURL.RawQuery)

	if err != nil {
		return "", err
	}

	queryStruct, err := query.Values(URLStruct)
	if err != nil {
		return "", err
	}

	for key, value := range queryStruct {
		for _, v := range value {
			parsedQuery.Add(key, v)
		}
	}

	return addQueryParams(parsedURL, parsedQuery), nil
}

func addQueryParams(parsedURL *url.URL, parsedQuery url.Values) string {
	return strings.Join([]string{strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1), parsedQuery.Encode()}, "?")
}

// Req takes 3 parameters and returns a Response struct.
// param method: HTTP method, "GET"/"POST"/"PUT"/"PATCH"/"DELETE"
// param    url: A URL
// param     ro: A RequestOptions struct
// If you do not intend to use the `RequestOptions` you can just pass nil
func (r RequestStruct) Req(method string, url string, ro *grequests.RequestOptions) (*grequests.Response, error) {
	if ro == nil {
		ro = &grequests.RequestOptions{}
	}

	var err error
	switch {
	case len(ro.Params) != 0:
		if url, err = buildURLParams(url, ro.Params); err != nil {
			return nil, err
		}
	case ro.QueryStruct != nil:
		if url, err = buildURLStruct(url, ro.QueryStruct); err != nil {
			return nil, err
		}
	}

	fullPath, err := getRequestURI(url)
	if err != nil {
		return nil, err
	}

	configs := r.configs
	ak := AuthKey{AccessKey: configs.Accesskey, SecretKey: configs.Secretkey}
	authKey := ak.Key(fullPath, method, 3600)
	if ro.Headers == nil {
		ro.Headers = map[string]string{}
	}
	ro.Headers["Authorization"] = authKey
	return grequests.DoRegularRequest(method, url, ro)
}

// Get takes 2 parameters and returns a Response struct.
// param    url: A URL
// param     ro: A RequestOptions struct
// If you do not intend to use the `RequestOptions` you can just pass nil
func (r RequestStruct) Get(url string, ro *grequests.RequestOptions) (*grequests.Response, error) {
	return r.Req("GET", url, ro)
}

// Post takes 2 parameters and returns a Response struct.
// param    url: A URL
// param     ro: A RequestOptions struct
// If you do not intend to use the `RequestOptions` you can just pass nil
func (r RequestStruct) Post(url string, ro *grequests.RequestOptions) (*grequests.Response, error) {
	return r.Req("POST", url, ro)
}

// Put takes 2 parameters and returns a Response struct.
// param    url: A URL
// param     ro: A RequestOptions struct
// If you do not intend to use the `RequestOptions` you can just pass nil
func (r RequestStruct) Put(url string, ro *grequests.RequestOptions) (*grequests.Response, error) {
	return r.Req("PUT", url, ro)
}

// Patch takes 2 parameters and returns a Response struct.
// param    url: A URL
// param     ro: A RequestOptions struct
// If you do not intend to use the `RequestOptions` you can just pass nil
func (r RequestStruct) Patch(url string, ro *grequests.RequestOptions) (*grequests.Response, error) {
	return r.Req("PATCH", url, ro)
}

// Delete takes 2 parameters and returns a Response struct.
// param    url: A URL
// param     ro: A RequestOptions struct
// If you do not intend to use the `RequestOptions` you can just pass nil
func (r RequestStruct) Delete(url string, ro *grequests.RequestOptions) (*grequests.Response, error) {
	return r.Req("DELETE", url, ro)
}

// Options takes 2 parameters and returns a Response struct.
// param    url: A URL
// param     ro: A RequestOptions struct
// If you do not intend to use the `RequestOptions` you can just pass nil
func (r RequestStruct) Options(url string, ro *grequests.RequestOptions) (*grequests.Response, error) {
	return r.Req("OPTIONS", url, ro)
}

// Head takes 2 parameters and returns a Response struct.
// param    url: A URL
// param     ro: A RequestOptions struct
// If you do not intend to use the `RequestOptions` you can just pass nil
func (r RequestStruct) Head(url string, ro *grequests.RequestOptions) (*grequests.Response, error) {
	return r.Req("HEAD", url, ro)
}
