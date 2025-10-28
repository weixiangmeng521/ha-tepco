# ARG BUILD_FROM
# FROM $BUILD_FROM

# # Install requirements for add-on
# RUN apk add --no-cache 
       
# # 安装 Chromium（Debian/Ubuntu）
# RUN apk add --no-cache \
#       chromium \
#       nss \
#       freetype \
#       ttf-freefont \
#       font-noto-cjk \ 
#       harfbuzz \
#       ca-certificates \
#       dumb-init

# # 让 go-rod 使用系统 Chromium
# ENV ROD_BROWSER_PATH=/usr/bin/chromium

# LABEL \
#     io.hass.version="VERSION" \
#     io.hass.type="addon" \
#     io.hass.arch="armhf|aarch64|i386|amd64"

# # 拷贝二进制和启动脚本
# COPY tepco-linux-aarch64 /usr/bin/tepco 
# COPY data/run.sh /
# RUN chmod +x /usr/bin/tepco /run.sh

# # 设置启动命令
# CMD [ "/run.sh" ]



ARG BUILD_FROM
FROM $BUILD_FROM

RUN echo "Building for architecture: $BUILD_ARCH"

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

LABEL \
    io.hass.version="VERSION" \
    io.hass.type="addon" \
    io.hass.arch="armhf|aarch64|i386|amd64"

# Execute during the build of the image
ARG APP_VERSION BUILD_ARCH

RUN echo "going to download bin file."

RUN \
    curl -L -o /usr/bin/tepco \
    "https://github.com/weixiangmeng521/ha-tepco/releases/download/v5.0.9/tepco-linux-${BUILD_ARCH}" \
    && chmod +x /usr/bin/tepco \
    && echo "✅ tepco downloaded" \
    && ls -lh /usr/bin/tepco

COPY data/run.sh /
RUN chmod +x /run.sh
# 默认运行 tepco
CMD [ "/run.sh" ]