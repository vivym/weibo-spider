FROM scratch

RUN mkdir /tmp

COPY weibo-cookie-renewer /

ENTRYPOINT ["/weibo-cookie-renewer"]
