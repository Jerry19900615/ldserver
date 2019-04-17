# ldserver
ldserver

学习用fcgi库，restful和fcgi的配合使用

用到的golang库:
fcgi + restful


nginx + fastcgi 
nginx配置：
server {
        listen 80;
        server_name go.dev;
        root /root/go/src/godev;
        index index.html;
        #gzip off;
        #proxy_buffering off;

        location / {
                 try_files $uri $uri/;
        }

        location ~ /app.* {
                include         fastcgi.conf;
                fastcgi_pass    127.0.0.1:9140;
        }

        try_files $uri $uri.html =404;
}


配置好后重启nginx

启动./bin/srv


