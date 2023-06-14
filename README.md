# pod-lifecycle-monitor


## Overview

pod lifecycle monitor is a Kubernetes controller used to monitor the lifecycle events of Pods in a Kubernetes cluster. It provides a simple way to track the state and progress of Pods as they are created, updated, and deleted, and is especially useful for debugging and troubleshooting

## Example

Here's an example of using Pod Lifecycle Monitor to monitor the lifecycle of Pods. This tool uses the Informer mechanism in Kubernetes to monitor the creation, update, and deletion of Pod resources.

### Step 1: Compile the Image Using Commands
```shell
git clone https://github.com/songkang7/pod-lifecycle-monitor.git
cd pod-lifecycle-monitor
docker build -t skto/pod-lifecycle-monitor:v1 . -f Dockerfile

```
### Step 2: Deploy Pod Lifecycle Monitor
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pod-life-cycle-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pod-life-cycle-cr
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: pod-life-cycle-rb
subjects:
  - kind: ServiceAccount
    name: pod-life-cycle-sa
    namespace: default
roleRef:
  kind: ClusterRole
  name: fake-time-injector-cr
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pod-life-cycle
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
        - name: pod-life-cycle
          image: skto/pod-lifecycle-monitor:v1
      serviceAccountName:  pod-life-cycle-sa
```

By default, logs will output information about Pods with a lifecycle of less than 30 seconds. You can adjust this by using the lifecycle parameter.
```yaml
...
spec:
  template:
    spec: 
      containers:
        - args:
            - '--lifecycle=60'               # Output information about Pods with a lifecycle of less than 60 seconds.
          name: pod-life-cycle
          image: skto/pod-lifecycle-monitor:v1
...
```

You can also view information about Pod creation and changes by setting the log level.

```yaml
...
spec:
  template:
    spec: 
      containers:
        - args:
            - '--v=2'               # When `v=2`, output Pod creation and change times.
          name: pod-life-cycle
          image: skto/pod-lifecycle-monitor:v1
...
```

## Conclusion

Pod Lifecycle Monitor provides a simple way to monitor the lifecycle of Pods in a Kubernetes cluster. By providing visibility into Pod state and progress, it makes it easier to debug and troubleshoot problems. With adjustable monitoring settings, you can tailor the tool to meet your specific needs.