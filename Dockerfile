FROM scratch

COPY tmp /tmp

COPY weibo-cookie-renewer /

ENTRYPOINT ["/weibo-cookie-renewer"]
