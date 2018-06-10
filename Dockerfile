FROM scratch
COPY bin/freenas-provisioner /
ENTRYPOINT ["/freenas-provisioner"]

ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt