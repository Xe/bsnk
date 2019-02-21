FROM xena/go:1.11.5 AS build

ENV GOPROXY=https://cache.greedo.xeserv.us
RUN GOBIN=/ go install

FROM xena/alpine
COPY --from=build /bsnk /usr/local/bin/bsnk

ENV PORT 5000
CMD /usr/local/bin/bsnk
