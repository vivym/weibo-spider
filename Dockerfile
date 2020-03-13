FROM scratch

COPY weibo-cookie-renewer /

ENTRYPOINT ["/weibo-cookie-renewer"]
