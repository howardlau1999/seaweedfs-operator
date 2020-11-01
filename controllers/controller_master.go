package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	seaweedv1 "github.com/seaweedfs/seaweedfs-operator/api/v1"
	label "github.com/seaweedfs/seaweedfs-operator/controllers/label"
)

func (r *SeaweedReconciler) ensureMaster(seaweedCR *seaweedv1.Seaweed) (done bool, result ctrl.Result, err error) {
	_ = context.Background()
	_ = r.Log.WithValues("seaweed", seaweedCR.Name)

	if done, result, err = r.ensureMasterPeerService(seaweedCR); done {
		return
	}

	if done, result, err = r.ensureMasterService(seaweedCR); done {
		return
	}

	if done, result, err = r.ensureMasterConfigMap(seaweedCR); done {
		return
	}

	if done, result, err = r.ensureMasterStatefulSet(seaweedCR); done {
		return
	}

	return
}

func (r *SeaweedReconciler) ensureMasterStatefulSet(seaweedCR *seaweedv1.Seaweed) (bool, ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("sw-master-statefulset", seaweedCR.Name)

	masterStatefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: seaweedCR.Name + "-master", Namespace: seaweedCR.Namespace}, masterStatefulSet)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.createMasterStatefulSet(seaweedCR)
		if err := controllerutil.SetControllerReference(seaweedCR, dep, r.Scheme); err != nil {
			return ReconcileResult(err)
		}
		log.Info("Creating a new master statefulset", "Namespace", dep.Namespace, "Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create master statefulset", "Namespace", dep.Namespace, "Name", dep.Name)
			return ReconcileResult(err)
		}

		// StatefulSet created successfully - return and requeue
		return ReconcileResult(err)
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ReconcileResult(err)
	}

	log.Info("master version " + masterStatefulSet.Spec.Template.Spec.Containers[0].Image + " expected " + seaweedCR.Spec.Image)
	if masterStatefulSet.Spec.Template.Spec.Containers[0].Image != seaweedCR.Spec.Image {
		masterStatefulSet.Spec.Template.Spec.Containers[0].Image = seaweedCR.Spec.Image
		if err = r.Update(ctx, masterStatefulSet); err != nil {
			log.Error(err, "Failed to update master statefulset", "Namespace", masterStatefulSet.Namespace, "Name", masterStatefulSet.Name)
			return ReconcileResult(err)
		}
		// Deployment created successfully - return and requeue
		return ReconcileResult(err)
	}

	log.Info("Get master stateful set " + masterStatefulSet.Name)
	return ReconcileResult(err)
}

func (r *SeaweedReconciler) ensureMasterConfigMap(seaweedCR *seaweedv1.Seaweed) (bool, ctrl.Result, error) {
	log := r.Log.WithValues("sw-master-configmap", seaweedCR.Name)

	masterConfigMap := r.createMasterConfigMap(seaweedCR)
	if err := controllerutil.SetControllerReference(seaweedCR, masterConfigMap, r.Scheme); err != nil {
		return ReconcileResult(err)
	}
	_, err := r.CreateOrUpdateConfigMap(masterConfigMap)

	log.Info("Get master ConfigMap " + masterConfigMap.Name)
	return ReconcileResult(err)
}

func (r *SeaweedReconciler) ensureMasterService(seaweedCR *seaweedv1.Seaweed) (bool, ctrl.Result, error) {
	log := r.Log.WithValues("sw-master-service", seaweedCR.Name)

	masterService := r.createMasterService(seaweedCR)
	if err := controllerutil.SetControllerReference(seaweedCR, masterService, r.Scheme); err != nil {
		return ReconcileResult(err)
	}
	_, err := r.CreateOrUpdateService(masterService)

	log.Info("Get master service " + masterService.Name)
	return ReconcileResult(err)
}

func (r *SeaweedReconciler) ensureMasterPeerService(seaweedCR *seaweedv1.Seaweed) (bool, ctrl.Result, error) {
	log := r.Log.WithValues("sw-master-peer-service", seaweedCR.Name)

	masterPeerService := r.createMasterPeerService(seaweedCR)
	if err := controllerutil.SetControllerReference(seaweedCR, masterPeerService, r.Scheme); err != nil {
		return ReconcileResult(err)
	}
	_, err := r.CreateOrUpdateService(masterPeerService)

	log.Info("Get master peer service " + masterPeerService.Name)
	return ReconcileResult(err)

}

func labelsForMaster(name string) map[string]string {
	return map[string]string{
		label.NameLabelKey:      "seaweedfs",
		label.ComponentLabelKey: "master",
		label.InstanceLabelKey:  name,
	}
}
