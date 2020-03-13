FROM chromedp/headless-shell:latest

COPY weibo-cookie-renewer /

ENTRYPOINT ["/weibo-cookie-renewer"]
