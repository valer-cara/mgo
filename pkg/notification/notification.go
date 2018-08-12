package notification

// Notification defines the interface that must be implemented by any notification provider
type Notification interface {
	// Sends a deployed notification
	Deployed(
		// repo is the repository that identifies the application
		repo string,
		// imageRepo is the link to the repository
		imageRepo string,
		// tag is the tag of the repository
		tag string,
		// cluster is where the deploy happened on
		cluster string,
		// author is the author's email
		author string,
		// error is the error if any
		err error,
	) error
}
