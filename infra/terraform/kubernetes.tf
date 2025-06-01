resource "kubernetes_deployment" "api_gateway" {
  metadata {
    name = "api-gateway"
    labels = {
      app = "api-gateway"
    }
  }

  spec {
    replicas = 3

    selector {
      match_labels = {
        app = "api-gateway"
      }
    }

    template {
      metadata {
        labels = {
          app = "api-gateway"
        }
      }

      spec {
        container {
          name  = "api-gateway"
          image = "api-gateway:latest"
          image_pull_policy = "Never"  # Для локального образа

          port {
            container_port = 8080
          }

          port {
            container_port = 9090  # Для метрик Prometheus
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "api_gateway_lb" {
  metadata {
    name = "api-gateway-lb"
  }
  spec {
    selector = {
      app = kubernetes_deployment.api_gateway.spec.0.template.0.metadata[0].labels.app
    }
    port {
      port        = 80
      target_port = 8080
    }
    type = "LoadBalancer"
  }
}

resource "kubernetes_ingress_v1" "api_gateway_ingress" {
  metadata {
    name = "api-gateway-ingress"
    annotations = {
      "nginx.ingress.kubernetes.io/rewrite-target" = "/"
    }
  }

  spec {
    ingress_class_name = "nginx"

    rule {
      host = "api-gateway.local"
      http {
        path {
          path = "/"
          backend {
            service {
              name = kubernetes_service.api_gateway_lb.metadata[0].name
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }
}