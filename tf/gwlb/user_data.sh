#!/bin/bash

# User data script for test EC2 instance
# This script sets up a simple web server to test traffic flow through the firewall

# Update system
dnf update -y

# Install required packages
dnf install -y httpd nginx openssl

# Generate self-signed certificate for HTTPS testing
mkdir -p /etc/ssl/private
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/ssl/private/server.key \
    -out /etc/ssl/certs/server.crt \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=${instance_name}"

# Configure Apache (HTTP on port 80)
cat > /var/www/html/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Firewall Test Instance</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 40px; 
            background-color: #f0f0f0; 
        }
        .container { 
            background-color: white; 
            padding: 20px; 
            border-radius: 10px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .status { color: #28a745; font-weight: bold; }
        .info { background-color: #e9ecef; padding: 10px; border-radius: 5px; margin: 10px 0; }
        .traffic-flow { background-color: #fff3cd; padding: 15px; border-radius: 5px; border-left: 4px solid #ffc107; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔥 Firewall Test Instance</h1>
        <p class="status">✅ HTTP Server is running on port 80</p>
        
        <div class="info">
            <h3>Instance Information:</h3>
            <ul>
                <li><strong>Instance Name:</strong> ${instance_name}</li>
                <li><strong>Local Time:</strong> <script>document.write(new Date().toLocaleString());</script></li>
                <li><strong>Server:</strong> Apache HTTP Server</li>
                <li><strong>Protocol:</strong> HTTP (Port 80)</li>
            </ul>
        </div>

        <div class="traffic-flow">
            <h3>🚦 Expected Traffic Flow:</h3>
            <ol>
                <li><strong>Internet</strong> → Network Load Balancer (Public Subnet)</li>
                <li><strong>NLB</strong> → Gateway Load Balancer Endpoint</li>
                <li><strong>GWLB</strong> → Network Firewall (Inspection)</li>
                <li><strong>Firewall</strong> → GWLB → Private Subnet</li>
                <li><strong>Private Instance</strong> → This web server</li>
            </ol>
            <p><em>All traffic between public and private subnets is inspected by the firewall!</em></p>
        </div>

        <div class="info">
            <h3>🧪 Test Commands:</h3>
            <pre>
# Test HTTP
curl http://[NLB-DNS-NAME]

# Test HTTPS  
curl -k https://[NLB-DNS-NAME]

# Check firewall logs in CloudWatch
            </pre>
        </div>
    </div>
</body>
</html>
EOF

# Start and enable Apache
systemctl start httpd
systemctl enable httpd

# Configure Nginx for HTTPS (port 443)
cat > /etc/nginx/nginx.conf << 'EOF'
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log;
pid /run/nginx.pid;

events {
    worker_connections 1024;
}

http {
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    server {
        listen 443 ssl;
        server_name _;
        
        ssl_certificate /etc/ssl/certs/server.crt;
        ssl_certificate_key /etc/ssl/private/server.key;
        
        ssl_session_cache shared:SSL:1m;
        ssl_session_timeout 5m;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        location / {
            root /usr/share/nginx/html;
            index index.html;
        }
    }
}
EOF

# Create HTTPS content
cat > /usr/share/nginx/html/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Firewall Test Instance - HTTPS</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 40px; 
            background-color: #f8f9fa; 
        }
        .container { 
            background-color: white; 
            padding: 20px; 
            border-radius: 10px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .status { color: #28a745; font-weight: bold; }
        .secure { color: #007bff; font-weight: bold; }
        .info { background-color: #e9ecef; padding: 10px; border-radius: 5px; margin: 10px 0; }
        .traffic-flow { background-color: #d1ecf1; padding: 15px; border-radius: 5px; border-left: 4px solid #17a2b8; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔒 Firewall Test Instance - HTTPS</h1>
        <p class="status">✅ HTTPS Server is running on port 443</p>
        <p class="secure">🔐 This connection is encrypted (self-signed certificate)</p>
        
        <div class="info">
            <h3>Instance Information:</h3>
            <ul>
                <li><strong>Instance Name:</strong> ${instance_name}</li>
                <li><strong>Local Time:</strong> <script>document.write(new Date().toLocaleString());</script></li>
                <li><strong>Server:</strong> Nginx</li>
                <li><strong>Protocol:</strong> HTTPS (Port 443)</li>
                <li><strong>SSL:</strong> Self-signed certificate</li>
            </ul>
        </div>

        <div class="traffic-flow">
            <h3>🔒 Secure Traffic Flow:</h3>
            <ol>
                <li><strong>Internet (HTTPS)</strong> → Network Load Balancer (Public Subnet)</li>
                <li><strong>NLB (TCP:443)</strong> → Gateway Load Balancer Endpoint</li>
                <li><strong>GWLB</strong> → Network Firewall (Deep Packet Inspection)</li>
                <li><strong>Firewall</strong> → GWLB → Private Subnet</li>
                <li><strong>Private Instance</strong> → This HTTPS server</li>
            </ol>
            <p><em>Even encrypted traffic metadata is inspected by the firewall!</em></p>
        </div>

        <div class="info">
            <h3>🧪 SSL Test Commands:</h3>
            <pre>
# Test HTTPS (accept self-signed cert)
curl -k https://[NLB-DNS-NAME]

# Test with certificate verification (will fail with self-signed)
curl https://[NLB-DNS-NAME]

# Check SSL certificate details
openssl s_client -connect [NLB-DNS-NAME]:443 -servername [NLB-DNS-NAME]
            </pre>
        </div>
    </div>
</body>
</html>
EOF

# Start and enable Nginx
systemctl start nginx
systemctl enable nginx

# Create a simple health check endpoint
mkdir -p /var/www/html/health
echo "OK" > /var/www/html/health/index.html

# Log the completion
echo "$(date): User data script completed successfully" >> /var/log/user-data.log
echo "HTTP server: http://localhost" >> /var/log/user-data.log
echo "HTTPS server: https://localhost" >> /var/log/user-data.log
