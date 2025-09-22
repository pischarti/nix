# nfw-tgw-spokes

## TF Helper

```sh
brew install kreuzwerker/taps/m1-terraform-provider-helper
m1-terraform-provider-helper activate
m1-terraform-provider-helper install hashicorp/template -v 2.2.0
```

## Tests

| Test Description | Expected Result | Test Type |
|------------------|-----------------|-----------|
| Ping between instances in `spoke-vpc-a` and `spoke-vpc-b` | ❌ Should **not** work | Inter-VPC Connectivity |
| SSH to EC2 instance in `spoke-vpc-b` from `spoke-vpc-a` (or vice-versa) | ❌ Should **not** work | Inter-VPC Connectivity |
| curl private IP of EC2 instance in `spoke-vpc-b` from `spoke-vpc-a` (or vice-versa) | ✅ Should work | Inter-VPC HTTP Traffic |
| Ping to a public IP address | ❌ Should **not** work | Internet Access (ICMP) |
| `dig` using a public DNS resolver | ❌ Should **not** work | DNS Filtering |
| curl `https://facebook.com` or `https://twitter.com` | ❌ Should **not** work | Domain Filtering |
| curl any other public URL | ✅ Should work | Internet Access (HTTP) |
| Navigate to `http://<public_alb_dns_name>` in browser | ✅ Should work | Public Load Balancer |

> **Note**: `<public_alb_dns_name>` is the DNS name of the ALB created in the Inspection VPC.

