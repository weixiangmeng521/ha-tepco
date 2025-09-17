ARG BUILD_FROM
FROM $BUILD_FROM

# Install requirements for add-on
RUN apk add --no-cache 
       
# 安装 Chromium（Debian/Ubuntu）
RUN apk add --no-cache \
      chromium \
      nss \
      freetype \
      ttf-freefont \
      font-noto-cjk \ 
      harfbuzz \
      ca-certificates \
      dumb-init

# 让 go-rod 使用系统 Chromium
ENV ROD_BROWSER_PATH=/usr/bin/chromium

LABEL \
    io.hass.version="VERSION" \
    io.hass.type="addon" \
    io.hass.arch="armhf|aarch64|i386|amd64"

# 拷贝二进制和启动脚本
COPY tepco-linux-aarch64 /usr/bin/tepco 
COPY data/run.sh /
RUN chmod +x /usr/bin/tepco /run.sh

# 设置启动命令
CMD [ "/run.sh" ]