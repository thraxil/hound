FROM centurylink/ca-certs
COPY hound /
COPY index.html /
COPY alert.html /
ENV HOUND_HTTP_PORT=9998
ENV HOUND_TEMPLATE_FILE=/index.html
ENV HOUND_SMTP_SERVER=postfix
ENV HOUND_SMTP_PORT=25
EXPOSE 9998
CMD ["/hound", "-config=/config.json"]
