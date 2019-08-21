FROM golang:1.12.8 as bd
RUN adduser --disabled-login appuser
WORKDIR /github.com/layer5io/meshery-app-mesh
ADD . .
RUN cd cmd; go build -ldflags="-w -s" -a -o /meshery-app-mesh .
RUN find . -name "*.go" -type f -delete; mv app-mesh /

FROM alpine
RUN apk --update add ca-certificates
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=bd /meshery-app-mesh /app/
COPY --from=bd /app-mesh /app/app-mesh
COPY --from=bd /etc/passwd /etc/passwd
USER appuser
WORKDIR /app
CMD ./meshery-app-mesh
