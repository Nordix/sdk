package interposecandidate

import (
	"bytes"
	"text/template"

	"github.com/networkservicemesh/api/pkg/api/registry"
)

// isSubset checks if B is a subset of A. TODO: reconsider this as a part of "tools"
func isSubset(a, b, nsLabels map[string]string) bool {
	if len(a) < len(b) {
		return false
	}
	for k, v := range b {
		if a[k] != v {
			result := ProcessLabels(v, nsLabels)
			if a[k] != result {
				return false
			}
		}
	}
	return true
}


// matchEndpoint return true if the endpoint is a feasible match based on the labels compared
func matchEndpoint(nsLabels map[string]string, nse *registry.NetworkServiceEndpoint) bool {
	// Iterate through network service labels
	// nsLabels must contain all the labels present in nse, and must match the values
	
	//note: do not report match in case nse has no network services
	if len(nse.GetNetworkServiceNames()) == 0 {
		return false
	}
	
	//note: do not report match in case nse has no labels for a particular network service name of his
	for _, v := range nse.GetNetworkServiceNames() {
		if nsls := nse.GetNetworkServiceLabels(); nsls != nil {
			if nsl, ok := nsls[v]; ok {
				if !isSubset(nsLabels, nsl.Labels, nsLabels) {
					return false
				}
			} else {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// ProcessLabels generates matches based on destination label selectors that specify templating.
func ProcessLabels(str string, vars interface{}) string {
	tmpl, err := template.New("tmpl").Parse(str)

	if err != nil {
		panic(err)
	}
	return process(tmpl, vars)
}

func process(t *template.Template, vars interface{}) string {
	var tmplBytes bytes.Buffer

	err := t.Execute(&tmplBytes, vars)
	if err != nil {
		panic(err)
	}
	return tmplBytes.String()
}
