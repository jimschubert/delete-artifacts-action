<<<<<<< Updated upstream
FROM gcr.io/distroless/static-debian12:nonroot
||||||| Stash base
FROM gcr.io/distroless/static-debian12
=======
FROM dhi.io/static:20250419
>>>>>>> Stashed changes
ARG APP_NAME
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/${APP_NAME} /app
ENTRYPOINT ["/app"]
