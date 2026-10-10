/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filter

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IncludeExclude struct {
	includes map[string]struct{}
	excludes map[string]struct{}
}

// DefaultExcludeResources defines cluster-scoped or ephemeral resources
// that should be excluded by default during backup or restore operations.
var DefaultExcludeResources = []string{
	"nodes",
	"endpointslices.discovery.k8s.io",
}

func NewIncludeExclude() *IncludeExclude {
	return &IncludeExclude{
		includes: make(map[string]struct{}),
		excludes: make(map[string]struct{}),
	}
}

func (f *IncludeExclude) Includes(items ...string) *IncludeExclude {
	for _, item := range items {
		f.includes[item] = struct{}{}
	}
	return f
}

func (f *IncludeExclude) Excludes(items ...string) *IncludeExclude {
	for _, item := range items {
		f.excludes[item] = struct{}{}
	}
	return f
}

func GetIncludeExcludeResources(includes, excludes []string) *IncludeExclude {
	f := NewIncludeExclude()
	f.Includes(includes...)
	for _, item := range excludes {
		if item == "*" {
			continue
		}
		f.Excludes(item)
	}
	return f
}

func (f *IncludeExclude) ShouldInclude(groupResource string) bool {
	// Always excluded if in excludes.

	if _, blocked := f.excludes[groupResource]; blocked {
		return false
	}
	// If no explicit includes or wildcard, include everything by default.
	if len(f.includes) == 0 || hasWildcard(f.includes) || exists(f.includes, groupResource) {
		return true
	}

	resource := getResourceFromGroupResource(groupResource)
	if _, blocked := f.excludes[resource]; blocked {
		return false
	}
	// If no explicit includes or wildcard, include everything by default.
	if len(f.includes) == 0 || hasWildcard(f.includes) || exists(f.includes, resource) {
		return true
	}

	return false
}

func getResourceFromGroupResource(groupResource string) string {
	parts := strings.Split(groupResource, ".")
	return parts[0]
}

func hasWildcard(set map[string]struct{}) bool {
	_, ok := set["*"]
	return ok
}

func exists(set map[string]struct{}, item string) bool {
	_, ok := set[item]
	return ok
}

type GlobalIncludeExclude struct {
	resourceFilter          *IncludeExclude
	includeClusterResources *bool
	namespaceFilter         *IncludeExclude
}

func NewGlobalIncludeExclude(resourceFilter, namespaceFilter *IncludeExclude, includeClusterResources bool) *GlobalIncludeExclude {
	return &GlobalIncludeExclude{
		resourceFilter:          resourceFilter,
		namespaceFilter:         namespaceFilter,
		includeClusterResources: ptr.To(includeClusterResources),
	}
}

func (g *GlobalIncludeExclude) ShouldIncludeResource(groupResource string, namespaced bool) bool {
	// If cluster-scoped and cluster resources not allowed, exclude.
	if !namespaced && !ptr.Deref(g.includeClusterResources, false) {
		return false
	}
	// Exclude default cluster-managed or ephemeral resources.
	if slices.Contains(DefaultExcludeResources, groupResource) {
		return false
	}
	// User-defined includes/excludes.
	return g.resourceFilter.ShouldInclude(groupResource)
}

func (g *GlobalIncludeExclude) ShouldIncludeNamespace(namespace string) bool {
	return g.namespaceFilter.ShouldInclude(namespace)
}

func IsPVCPassedTheFilter(object client.Object, params *runtime.RawExtension) bool {
	strParams := make(map[string]string)
	if params != nil && params.Raw != nil {
		if err := json.Unmarshal(params.Raw, &strParams); err != nil {
			klog.Warningf("failed to unmarshal filter parameters: %v", err)
			return true
		}
	}

	IncludeResources := trimSpaceFromList(strParams["IncludeResources"])
	IncludeNamespaces := trimSpaceFromList(strParams["IncludeNamespaces"])
	ExcludeNamespaces := trimSpaceFromList(strParams["ExcludeNamespaces"])
	ExcludeResources := trimSpaceFromList(strParams["ExcludeResources"])
	ORedLabelSelectors := trimSpaceFromList(strParams["ORedLabelSelectors"])
	ANDedLabelSelectors := trimSpaceFromList(strParams["ANDedLabelSelectors"])

	resFilter := NewIncludeExclude().Includes(IncludeResources...).Excludes(ExcludeResources...)
	nsFilter := NewIncludeExclude().Includes(IncludeNamespaces...).Excludes(ExcludeNamespaces...)
	globalFilter := NewGlobalIncludeExclude(resFilter, nsFilter, false)

	if isFilteredOutByLabelSelectors(object, ORedLabelSelectors, ANDedLabelSelectors) {
		return false
	}
	groupResource := fmt.Sprintf("%s.%s", "persistentvolumeclaims", "")
	if !globalFilter.ShouldIncludeResource(groupResource, true) {
		return false
	}
	return true
}

func isFilteredOutByLabelSelectors(object client.Object, ORedLabelSelectors, ANDedLabelSelectors []string) bool {
	labels := labelsToStrings(object.GetLabels())

	if !matchesAny(labels, ORedLabelSelectors) {
		klog.Infof("no OR labels match, skipped\n")
		return false
	}
	if !matchesAll(labels, ANDedLabelSelectors) {
		klog.Infof("not all AND labels match, skipped\n")
		return false
	}
	return true
}

func labelsToStrings(labels map[string]string) []string {
	out := make([]string, 0, len(labels))
	for k, v := range labels {
		out = append(out, fmt.Sprintf("%s:%s", k, v))
		out = append(out, fmt.Sprintf("%s=%s", k, v))
		out = append(out, k)
	}
	return out
}

func matchesAny(labels, selectors []string) bool {
	if len(selectors) == 0 {
		return true
	}
	set := sets.NewString(labels...)
	for _, sel := range selectors {
		if set.Has(sel) {
			return true
		}
	}
	return false
}

func matchesAll(labels, selectors []string) bool {
	if len(selectors) == 0 {
		return true
	}
	set := sets.NewString(labels...)
	for _, sel := range selectors {
		if !set.Has(sel) {
			return false
		}
	}
	return true
}

func trimSpaceFromList(input string) []string {
	inputs := strings.Split(input, ",")
	for i, val := range inputs {
		inputs[i] = strings.TrimSpace(val)
	}
	return inputs
}
