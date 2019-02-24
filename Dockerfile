FROM xena/go:1.11.5 AS build

ENV GOPROXY=https://cache.greedo.xeserv.us
WORKDIR /bsnk
COPY . .
RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install ./cmd/bsnk

FROM xena/alpine
COPY --from=build /usr/local/bin/bsnk /usr/local/bin/bsnk

ENV PORT 5000
CMD /usr/local/bin/bsnk
