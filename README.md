# k8s-resource-tracker (WIP)

[![Build](https://github.com/sfotiadis/k8s-resource-tracker/actions/workflows/go.yml/badge.svg)](https://github.com/sfotiadis/k8s-resource-tracker/actions/workflows/go.yml)
[![Vulnerability Check](https://github.com/sfotiadis/k8s-resource-tracker/actions/workflows/vulncheck.yml/badge.svg)](https://github.com/sfotiadis/k8s-resource-tracker/actions/workflows/vulncheck.yml)

Welcome to the k8s-resource-tracker project! This utility monitors the resource usage of Kubernetes pods and provides insights into CPU and memory consumption. Keep track of your pod's performance effortlessly.

## Table of Contents

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Example Usage](#example-usage)
- [Deployment](#deployment)
- [Customization](#customization)
- [Contributing](#contributing)
- [License](#license)

## Introduction

The **k8s-resource-tracker** utility is designed to monitor resource usage, specifically CPU and memory consumption, of Kubernetes pods. It utilizes the Prometheus client library to expose metrics and enable monitoring through Prometheus and other compatible monitoring systems.

## Prerequisites

- A Kubernetes cluster
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and configured
- [Helm](https://helm.sh/docs/intro/install/) installed (for Prometheus deployment)
- Basic understanding of Kubernetes and Prometheus

## Installation

1. Clone the repository to your local machine:
   ```bash
   git clone https://github.com/sfotiadis/k8s-resource-tracker.git
   ```

2. Navigate to the project directory:
   ```bash
   cd k8s-resource-tracker
   ```

## Usage

1. Update the `deployment.yaml` file to configure the namespace, label selector, and other settings.
2. Apply the deployment:
   ```bash
   kubectl apply -f deployment.yaml
   ```

## Example Usage

Here's an example of how to use the `k8s-resource-tracker`:

1. Deploy Prometheus using Helm:
   ```bash
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm install prometheus prometheus-community/kube-prometheus-stack
   ```

2. Apply the `deployment.yaml` manifest:
   ```bash
   kubectl apply -f deployment.yaml
   ```

3. Access the Prometheus UI to visualize the exposed metrics:
   ```bash
   kubectl --namespace default port-forward svc/prometheus-kube-prometheus-prometheus 9090
   ```

4. Open a web browser and navigate to http://localhost:9090 to access the Prometheus UI.

## Deployment

1. Navigate to the cmd directory:
   ```bash
   cd k8s-resource-tracker/cmd/
   ```

1. **Create a Docker Image:**

   Create a Dockerfile in the same directory as the k8s-resource-tracker.go file with the following contents:

   ```Dockerfile
   # Use an official Golang runtime as the base image
   FROM golang:1.21

   # Set the working directory to the app directory
   WORKDIR /app

   # Copy the current directory contents into the container at /app
   COPY . /app

   # Build the app
   RUN go build -o resource-tracker .

   # Set the entry point of the container to the app
   ENTRYPOINT ["./resource-tracker"]
   ```

   Then, build the Docker image:

   ```bash
   docker build -t resource-tracker .
   ```

2. **Push Docker Image:**

   Push the Docker image to a container registry (like Docker Hub or your organization's registry):

   ```bash
   docker tag resource-tracker yourregistry/resource-tracker:v1.0
   docker push yourregistry/resource-tracker:v1.0
   ```

3. **Create Kubernetes Deployment:**

   Create a Kubernetes Deployment YAML file (e.g., `resource-tracker-deployment.yaml`) with the following content:

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: resource-tracker
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: resource-tracker
     template:
       metadata:
         labels:
           app: resource-tracker
       spec:
         containers:
           - name: resource-tracker
             image: yourregistry/resource-tracker:v1.0
             command: ["./resource-tracker"]
             args: ["-namespace=default", "-pod-label=app=myapp"]
             resources:
               requests:
                 cpu: "200m"
                 memory: "256Mi"
               limits:
                 cpu: "200m"
                 memory: "256Mi"
   ```

   Replace `yourregistry` with the appropriate registry URL.


4. **Apply the Deployment:**

   Apply the deployment to your Kubernetes cluster:

   ```bash
   kubectl apply -f resource-tracker-deployment.yaml
   ```

## Customization

Feel free to customize the `deployment.yaml` file according to your monitoring needs. You can adjust the namespace, label selector, monitoring interval, and other settings.

## Contributing

Contributions to this project are welcome! If you'd like to contribute, please follow the guidelines mentioned in the [Contributing](CONTRIBUTING.md) document.

## License

This project is licensed under the [MIT License](LICENSE).

---

Thank you for exploring the k8s-resource-tracker project. Keep your Kubernetes pods' resource usage in check with this monitoring utility!
