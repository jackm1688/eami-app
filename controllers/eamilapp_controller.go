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

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	emailappv1 "email-app/api/v1"
)

// EamilAppReconciler reconciles a EamilApp object
type EamilAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=emailapp.lemon.cn,resources=eamilapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=emailapp.lemon.cn,resources=eamilapps/status,verbs=get;update;patch

func (r *EamilAppReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("eamilapp", req.NamespacedName)

	// your logic here
	log.Info("start work")
	cr := &emailappv1.EamilApp{}
	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("not found emailApp CR,maybe removed")
			return ctrl.Result{}, nil
		}
		log.Error(err, "get emailApp CR failed")
		return ctrl.Result{}, err
	}

	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      cr.Spec.AppName,
	}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			if *cr.Spec.TotalQPS < 1 {
				log.Info("not need to create deployment,because total qpl is 0")
				return ctrl.Result{}, nil
			}
			//create service
			log.Info("create svc resource")
			if err := createService(ctx, r, cr, req); err != nil {
				log.Error(err, "create svc resource failed")
				return ctrl.Result{}, err
			}
			//create deployment resource
			log.Info("create deployment resource")
			if err := createDeployment(ctx, r, cr, req); err != nil {
				log.Error(err, "create deploy resource failed")
				return ctrl.Result{}, err
			}

			//update status
			if err := updateStatus(ctx, r, cr); err != nil {
				log.Error(err, "update emailApp cr status failed")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "get deployment resource failed")
		return ctrl.Result{}, err
	}

	//update deploy
	if err := updateDeployment(ctx, r, cr, deployment); err != nil {
		log.Error(err, "update deployment obj failed")
		return ctrl.Result{}, err
	}
	//update status
	if err := updateStatus(ctx, r, cr); err != nil {
		log.Error(err, "update cr status  failed")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *EamilAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&emailappv1.EamilApp{}).
		Complete(r)
}
