package controllers

import (
	"context"
	emailappv1 "email-app/api/v1"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getExpectReplica(emailApp *emailappv1.EamilApp) int32 {

	singlePodQPS := *(emailApp.Spec.SinglePodQPS)
	totalQPS := *(emailApp.Spec.TotalQPS)
	replicas := totalQPS / singlePodQPS
	if totalQPS%singlePodQPS > 1 {
		replicas++
	}
	return replicas
}

func createService(ctx context.Context, r *EamilAppReconciler, cr *emailappv1.EamilApp, req ctrl.Request) error {
	log := r.Log.WithValues("func", "createService")
	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      cr.Spec.AppName + "-svc",
	}, svc)

	if errors.IsNotFound(err) {
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Spec.AppName + "-svc",
				Namespace: cr.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": cr.Spec.AppName,
				},
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{
						Name:     "http",
						Protocol: corev1.ProtocolTCP,
						Port:     *cr.Spec.ContainerPort,
					},
				},
			},
		}

		//build reference to controller
		if err := controllerutil.SetControllerReference(cr, svc, r.Scheme); err != nil {
			return err
		}
		log.Info("build reference to controller successfully")

		//create svc resource
		if err := r.Create(ctx, svc); err != nil {
			return err
		}
		log.Info("create svc resource successfully")
		return nil
	}
	return err
}

func createDeployment(ctx context.Context, r *EamilAppReconciler, cr *emailappv1.EamilApp, req ctrl.Request) error {

	log := r.Log.WithValues("func", "createDeployment")
	expectReplicas := getExpectReplica(cr)
	log.Info(fmt.Sprintf("expectReplicas :%d", expectReplicas))

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cr.Spec.AppName,
				},
			},
			Replicas: pointer.Int32Ptr(expectReplicas),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": cr.Spec.AppName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cr.Spec.AppName,
							Image: cr.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: *cr.Spec.ContainerPort,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu":    resource.MustParse(cr.Spec.CpuRequest),
									"memory": resource.MustParse(cr.Spec.MemRequest),
								},
								Requests: corev1.ResourceList{
									"cpu":    resource.MustParse(cr.Spec.CpuLimit),
									"memory": resource.MustParse(cr.Spec.MemLimit),
								},
							},
						},
					},
				},
			},
		},
	}
	//build reference to controller
	if err := controllerutil.SetControllerReference(cr, deployment, r.Scheme); err != nil {
		return err
	}
	log.Info("build reference to controller success")

	//create deployment resource
	if err := r.Create(ctx, deployment); err != nil {
		return err
	}
	log.Info("create deployment resource success")
	return nil
}

func updateStatus(ctx context.Context, r *EamilAppReconciler, cr *emailappv1.EamilApp) error {

	log := r.Log.WithValues("func", "updateStatus")
	singlePodQPS := *(cr.Spec.SinglePodQPS)
	expectReplicas := getExpectReplica(cr)

	if nil == cr.Status.RealQPS {
		cr.Status.RealQPS = new(int32)
	}

	*cr.Status.RealQPS = singlePodQPS * expectReplicas

	if err := r.Update(ctx, cr); err != nil {
		return err
	}
	log.Info("update emailAPp cr's realQPS success")
	return nil
}

func updateDeployment(ctx context.Context, r *EamilAppReconciler, cr *emailappv1.EamilApp, d *appsv1.Deployment) error {

	log := r.Log.WithValues("func", "updateDeployment")

	expectReplicas := getExpectReplica(cr)
	currentReplicas := *d.Spec.Replicas

	expectImage := cr.Spec.Image
	currentImage := d.Spec.Template.Spec.Containers[0].Image

	expectCPUR := resource.MustParse(cr.Spec.CpuRequest)
	currentCpuR := d.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()

	expectCPUL := resource.MustParse(cr.Spec.CpuLimit)
	currentCpuL := d.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()

	expectMEMR := resource.MustParse(cr.Spec.MemRequest)
	currentMEMR := d.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()

	expectMEML := resource.MustParse(cr.Spec.MemLimit)
	currentMEML := d.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()

	if expectReplicas == currentReplicas && expectImage == currentImage &&
		reflect.DeepEqual(expectCPUR, currentCpuR) && reflect.DeepEqual(expectCPUL, currentCpuL) &&
		reflect.DeepEqual(expectMEMR, currentMEMR) && reflect.DeepEqual(expectMEML, currentMEML) {

		log.Info("no changed,skip update")
		return nil
	}

	if expectReplicas != currentReplicas {
		*d.Spec.Replicas = expectReplicas
	}

	if expectImage != currentImage {
		d.Spec.Template.Spec.Containers[0].Image = expectImage
	}

	if !reflect.DeepEqual(expectCPUR, currentCpuR) || !reflect.DeepEqual(expectCPUL, currentCpuL) ||
		!reflect.DeepEqual(expectMEMR, currentMEMR) || !reflect.DeepEqual(expectMEML, currentMEML) {

		resources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"cpu":    resource.MustParse(cr.Spec.CpuRequest),
				"memory": resource.MustParse(cr.Spec.MemRequest),
			},
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse(cr.Spec.CpuLimit),
				"memory": resource.MustParse(cr.Spec.MemLimit),
			},
		}
		d.Spec.Template.Spec.Containers[0].Resources = resources
	}

	if err := r.Update(ctx, d); err != nil {
		return err
	}
	log.Info("update deployment obj success")
	return nil
}
