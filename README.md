# Kubernetes SQS Operator

A Kubernetes operator that simplifies the management of AWS Simple Queue Service (SQS) queues through Kubernetes custom resources. This operator allows you to declaratively create, configure, and manage SQS queues directly from your Kubernetes cluster.

## Features

- **Declarative Queue Management**: Define SQS queues using Kubernetes custom resources
- **Automatic Reconciliation**: The operator continuously ensures the desired state matches the actual AWS SQS queue state
- **Support for FIFO and Standard Queues**: Create both FIFO (first-in, first-out) and standard queues
- **Comprehensive Configuration**: Configure visibility timeout, message retention, delay seconds, max message size, and tags
- **Lifecycle Management**: Handles queue creation, updates, and deletion with proper cleanup
- **Status Reporting**: Provides real-time status of queue existence and configuration

## Prerequisites

- Kubernetes cluster (v1.16+)
- AWS account with appropriate permissions to manage SQS queues
- AWS credentials configured (IAM user/role with SQS permissions)
- Go 1.23+ (for development)

### AWS Permissions

The operator requires the following IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:CreateQueue",
        "sqs:GetQueueAttributes",
        "sqs:SetQueueAttributes",
        "sqs:DeleteQueue",
        "sqs:ListQueues",
        "sqs:GetQueueUrl"
      ],
      "Resource": "*"
    }
  ]
}
```

## Installation

### Deploy to Kubernetes

1. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/kubernetes-sqs-operator.git
   cd kubernetes-sqs-operator
   ```

2. Build the operator:
   ```bash
   go build -o bin/manager cmd/main.go
   ```

3. Build and push the Docker image:
   ```bash
   docker build -t your-registry/kubernetes-sqs-operator:v0.1.0 .
   docker push your-registry/kubernetes-sqs-operator:v0.1.0
   ```

4. Deploy the CRDs and operator to your cluster (you'll need to create deployment manifests or use a tool like Kustomize/Helm)

### Local Development

1. Install the CRDs (if using kubebuilder):
   ```bash
   # If you have kubebuilder installed
   kubebuilder create api --group messaging --version v1 --kind SqsQueue
   ```

2. Run the operator locally:
   ```bash
   go run cmd/main.go
   ```

## Usage

### Creating an SQS Queue

Create a YAML file for your SQS queue:

```yaml
apiVersion: messaging.cloud.com/v1
kind: SqsQueue
metadata:
  name: my-queue
  namespace: default
spec:
  queueName: my-application-queue
  region: us-east-1
  fifo: false
  visibilityTimeoutSeconds: 30
  messageRetentionSeconds: 345600
  delaySeconds: 0
  maxMessageSizeBytes: 262144
  tags:
    Environment: production
    Application: my-app
```

Apply the resource:

```bash
kubectl apply -f my-queue.yaml
```

### FIFO Queue Example

```yaml
apiVersion: messaging.cloud.com/v1
kind: SqsQueue
metadata:
  name: my-fifo-queue
  namespace: default
spec:
  queueName: my-application-queue.fifo
  region: us-east-1
  fifo: true
  visibilityTimeoutSeconds: 60
  messageRetentionSeconds: 604800
```

### Checking Queue Status

View the status of your queues:

```bash
kubectl get sqsqueues
```

Example output:
```
NAME           QUEUENAME               STATE    FIFO   QUEUEURL
my-queue       my-application-queue    Active   false  https://sqs.us-east-1.amazonaws.com/123456789012/my-application-queue
my-fifo-queue  my-application-queue.fifo Active true   https://sqs.us-east-1.amazonaws.com/123456789012/my-application-queue.fifo
```

### Deleting a Queue

Delete the Kubernetes resource, and the operator will automatically delete the SQS queue:

```bash
kubectl delete sqsqueue my-queue
```

## Configuration Options

| Field | Type | Description | Default | Valid Range |
|-------|------|-------------|---------|-------------|
| `queueName` | string | Name of the SQS queue | Required | - |
| `region` | string | AWS region | Required | - |
| `fifo` | bool | Whether to create a FIFO queue | false | - |
| `visibilityTimeoutSeconds` | int32 | Message visibility timeout | 30 | 0-43200 |
| `messageRetentionSeconds` | int32 | How long to retain messages | 345600 | 60-1209600 |
| `delaySeconds` | int32 | Delivery delay for all messages | 0 | 0-900 |
| `maxMessageSizeBytes` | int32 | Maximum message size | 262144 | 1024-262144 |
| `tags` | map[string]string | AWS resource tags | - | - |

## Development

### Building

```bash
go build -o bin/manager cmd/main.go
```

### Testing

```bash
go test ./...
```

### Code Generation

If using kubebuilder for code generation:

```bash
# Generate deepcopy methods
controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Generate CRDs
controller-gen crd:trivialVersions=true paths="./..." output:crd:artifacts:config=config/crd/bases
```

## Architecture

The operator consists of:

- **Custom Resource Definition (CRD)**: Defines the `SqsQueue` custom resource
- **Controller**: Watches for `SqsQueue` resources and reconciles their state with AWS SQS
- **AWS Client**: Handles all interactions with AWS SQS API

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have questions:

- Open an issue on GitHub
- Check the [troubleshooting guide](docs/troubleshooting.md) (if available)
- Review the AWS SQS documentation

## Related Projects

- [AWS SQS Documentation](https://docs.aws.amazon.com/sqs/)
- [Kubernetes Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [Controller Runtime](https://github.com/kubernetes-sigs/controller-runtime)</content>
<parameter name="filePath">c:\Users\Admin\kubernetes-sqs-operator\README.md
