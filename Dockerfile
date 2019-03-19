FROM xena/go:1.12.1 AS build
ENV GOPROXY=https://cache.greedo.xeserv.us
WORKDIR /bsnk
COPY . .
RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install ./cmd/bsnk

FROM xena/alpine
COPY ./app /app
COPY --from=build /usr/local/bin/bsnk /usr/local/bin/bsnk
ENV PORT 5000
CMD /usr/local/bin/bsnk
