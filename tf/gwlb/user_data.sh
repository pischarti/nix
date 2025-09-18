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
error_log /var/log/nginx/error.log warn;
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
    server_tokens off;

    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # HTTPS Server
    server {
        listen 443 ssl http2;
        server_name _;
        
        # SSL Configuration
        ssl_certificate /etc/ssl/certs/server.crt;
        ssl_certificate_key /etc/ssl/private/server.key;
        
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
        ssl_prefer_server_ciphers off;
        ssl_session_cache shared:SSL:10m;
        ssl_session_timeout 10m;
        ssl_session_tickets off;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";

        # Document root
        root /usr/share/nginx/html;
        index index.html;

        # Main location
        location / {
            try_files $uri $uri/ =404;
        }

        # Health check endpoint for HTTPS
        location /health {
            access_log off;
            return 200 "OK\n";
            add_header Content-Type text/plain;
        }

        # Status endpoint
        location /status {
            access_log off;
            return 200 "HTTPS Server Running\n";
            add_header Content-Type text/plain;
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

# Test Nginx configuration before starting
nginx -t
if [ $? -ne 0 ]; then
    echo "$(date): ERROR - Nginx configuration test failed" >> /var/log/user-data.log
    systemctl status nginx >> /var/log/user-data.log 2>&1
    exit 1
fi

# Start and enable Nginx
systemctl start nginx
systemctl enable nginx

# Wait for Nginx to start and verify
sleep 5
if systemctl is-active --quiet nginx; then
    echo "$(date): SUCCESS - Nginx started successfully" >> /var/log/user-data.log
else
    echo "$(date): ERROR - Nginx failed to start" >> /var/log/user-data.log
    systemctl status nginx >> /var/log/user-data.log 2>&1
    journalctl -u nginx --no-pager >> /var/log/user-data.log 2>&1
fi

# Create a simple health check endpoint for Apache
mkdir -p /var/www/html/health
echo "OK" > /var/www/html/health/index.html

# Verify both servers are responding
echo "$(date): Testing local HTTP server..." >> /var/log/user-data.log
if curl -s http://localhost/health > /dev/null; then
    echo "$(date): SUCCESS - HTTP health check passed" >> /var/log/user-data.log
else
    echo "$(date): ERROR - HTTP health check failed" >> /var/log/user-data.log
fi

echo "$(date): Testing local HTTPS server..." >> /var/log/user-data.log
if curl -k -s https://localhost/health > /dev/null; then
    echo "$(date): SUCCESS - HTTPS health check passed" >> /var/log/user-data.log
else
    echo "$(date): ERROR - HTTPS health check failed" >> /var/log/user-data.log
fi

# Check listening ports
echo "$(date): Checking listening ports..." >> /var/log/user-data.log
netstat -tlnp | grep -E ':(80|443) ' >> /var/log/user-data.log 2>&1

# Log the completion
echo "$(date): User data script completed successfully" >> /var/log/user-data.log
echo "HTTP server: http://localhost" >> /var/log/user-data.log
echo "HTTPS server: https://localhost" >> /var/log/user-data.log
echo "Logs available at: /var/log/user-data.log" >> /var/log/user-data.log

# Create a status script for troubleshooting
cat > /home/ec2-user/check_servers.sh << 'EOF'
#!/bin/bash
echo "=== Server Status Check ==="
echo "Date: $(date)"
echo ""

echo "=== Service Status ==="
systemctl status httpd --no-pager
echo ""
systemctl status nginx --no-pager
echo ""

echo "=== Listening Ports ==="
netstat -tlnp | grep -E ':(80|443) '
echo ""

echo "=== Health Checks ==="
echo -n "HTTP Health Check: "
if curl -s http://localhost/health > /dev/null; then
    echo "PASS"
else
    echo "FAIL"
fi

echo -n "HTTPS Health Check: "
if curl -k -s https://localhost/health > /dev/null; then
    echo "PASS"
else
    echo "FAIL"
fi

echo ""
echo "=== Recent Logs ==="
echo "--- User Data Log (last 10 lines) ---"
tail -10 /var/log/user-data.log
echo ""
echo "--- Nginx Error Log (last 10 lines) ---"
tail -10 /var/log/nginx/error.log
EOF

chmod +x /home/ec2-user/check_servers.sh
chown ec2-user:ec2-user /home/ec2-user/check_servers.sh
