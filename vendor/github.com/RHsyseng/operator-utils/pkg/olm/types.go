package olm

type DeploymentStatus struct {
	// Deployments are ready to serve requests
	Ready []string `json:"ready,omitempty"`
	// Deployments are starting, may or may not succeed
	Starting []string `json:"starting,omitempty"`
	// Deployments are not starting, unclear what next step will be
	Stopped []string `json:"stopped,omitempty"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeploymentStatus) DeepCopyInto(out *DeploymentStatus) {
	*out = *in
	if in.Ready != nil {
		in, out := &in.Ready, &out.Ready
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Starting != nil {
		in, out := &in.Starting, &out.Starting
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Stopped != nil {
		in, out := &in.Stopped, &out.Stopped
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Deployments.
func (in *DeploymentStatus) DeepCopy() *DeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(DeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

type deployments interface {
	count() int
	name(i int) string
	requestedReplicas(i int) int32
	targetReplicas(i int) int32
	readyReplicas(i int) int32
}

type deploymentsWrapper struct {
	countFunc             func() int
	nameFunc              func(i int) string
	requestedReplicasFunc func(i int) int32
	targetReplicasFunc    func(i int) int32
	readyReplicasFunc     func(i int) int32
}

func (obj deploymentsWrapper) count() int {
	return obj.countFunc()
}

func (obj deploymentsWrapper) name(i int) string {
	return obj.nameFunc(i)
}

func (obj deploymentsWrapper) requestedReplicas(i int) int32 {
	return obj.requestedReplicasFunc(i)
}

func (obj deploymentsWrapper) targetReplicas(i int) int32 {
	return obj.targetReplicasFunc(i)
}

func (obj deploymentsWrapper) readyReplicas(i int) int32 {
	return obj.readyReplicasFunc(i)
}
