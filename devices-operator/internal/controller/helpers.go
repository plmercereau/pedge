package controller

import (
	"crypto/rand"
	"math/big"

	batchv1 "k8s.io/api/batch/v1"
)

func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
		"!@#$%^&*()-_=+[]{}|;:,.<>?/~"
	password := make([]byte, length)
	for i := range password {
		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[randomIndex.Int64()]
	}
	return string(password)
}

// jobSpecMatches checks if two job specs match
func jobSpecMatches(existingJob, newJob *batchv1.Job) bool {
	// Add your comparison logic here, comparing fields in existingJob.Spec and newJob.Spec
	// Return true if they match, false otherwise

	// For simplicity, we just check labels and some key fields here
	// TODO implement a better way to compare job specs
	if !equalMaps(existingJob.Labels, newJob.Labels) ||
		// (existingJob.Annotations[secretHashAnnotation] != newJob.Annotations[secretHashAnnotation]) ||
		existingJob.Spec.Template.Spec.RestartPolicy != newJob.Spec.Template.Spec.RestartPolicy ||
		len(existingJob.Spec.Template.Spec.InitContainers) != len(newJob.Spec.Template.Spec.InitContainers) ||
		len(existingJob.Spec.Template.Spec.Containers) != len(newJob.Spec.Template.Spec.Containers) {
		return false
	}

	// Add more comparison logic as needed

	return true
}

// equalMaps checks if two maps are equal
func equalMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
