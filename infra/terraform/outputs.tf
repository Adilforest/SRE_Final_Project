output "load_balancer_hostname" {
  value = kubernetes_ingress_v1.api_gateway_ingress.spec.0.rule.0.host
}

output "service_name" {
  value = kubernetes_service.api_gateway_lb.metadata[0].name
}