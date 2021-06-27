/*

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

package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// EamilAppSpec defines the desired state of EamilApp
type EamilAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of EamilApp. Edit EamilApp_types.go to remove/update
	AppName       string `json:"appName"`
	Image         string `json:"image"`
	ContainerPort *int32 `json:"containerPort"`
	SinglePodQPS  *int32 `json:"singlePodQPS"`
	TotalQPS      *int32 `json:"totalQPS"`
	CpuRequest    string `json:"cpuRequest"`
	CpuLimit      string `json:"cpuLimit"`
	MemRequest    string `json:"memRequest"`
	MemLimit      string `json:"memLimit"`
}

// EamilAppStatus defines the observed state of EamilApp
// +kubebuilder:subresource:status
type EamilAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	RealQPS *int32 `json:"realQPS"`
}

func (in *EamilApp) String() string {
	return fmt.Sprintf("AppName: %s,Image: %s, Container: %d,SinglePodQS: %d, TotalQPS: %d", in.Spec.AppName,
		in.Spec.Image, *in.Spec.ContainerPort, *in.Spec.SinglePodQPS, *in.Spec.TotalQPS)
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Deploy-name",type="string",JSONPath=".spec.appName",description="AppName of EamilApp"
// +kubebuilder:printcolumn:name="SinglePodQPS",type="integer",JSONPath=".spec.singlePodQPS",description="SinglePodQPS of EamilApp"
// +kubebuilder:printcolumn:name="TotalQPS",type="integer",JSONPath=".spec.totalQPS",description="TotalQPS of EamilApp"
// +kubebuilder:printcolumn:name="CpuRequest",type="string",JSONPath=".spec.cpuRequest",description="CpuRequest of EamilApp"
// +kubebuilder:printcolumn:name="CpuLimit",type="string",JSONPath=".spec.cpuLimit",description="CpuLimit of EamilApp"
// +kubebuilder:printcolumn:name="MemRequest",type="string",JSONPath=".spec.memRequest",description="MemRequest of EamilApp"
// +kubebuilder:printcolumn:name="MemLimit",type="string",JSONPath=".spec.memLimit",description="MemLimit of EamilApp"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// EamilApp is the Schema for the eamilapps API
type EamilApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EamilAppSpec   `json:"spec,omitempty"`
	Status EamilAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EamilAppList contains a list of EamilApp
type EamilAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EamilApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EamilApp{}, &EamilAppList{})
}
