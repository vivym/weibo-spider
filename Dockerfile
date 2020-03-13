FROM scratch

COPY weibo-spider /

ENTRYPOINT ["/weibo-spider"]
