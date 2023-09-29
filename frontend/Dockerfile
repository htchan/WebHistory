FROM plugfox/flutter:stable-web

WORKDIR /usr/src/app

RUN mkdir /build-result
RUN echo '{ "build-dir": "../../../build-result" }' > /root/.flutter_settings

CMD ['flutter pub get ; flutter build web']
