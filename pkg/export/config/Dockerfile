FROM debian:bullseye-slim AS base

# Largely stolen from:
# https://github.com/abarichello/godot-ci/

ENV DEBIAN_FRONTEND=noninteractive
ENV ANDROID_HOME=/opt/android-sdk

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	apt-get update && apt-get install --no-install-recommends -y \
	ca-certificates \
	openjdk-17-jdk-headless \
	gnupg \
	wget \
	unzip \
	zip

# ---

FROM base AS android

RUN mkdir /opt/android-sdk

RUN wget https://dl.google.com/android/repository/commandlinetools-linux-7583922_latest.zip \
	&& unzip commandlinetools-*.zip -d "${ANDROID_HOME}/cmdline-tools" \
	&& rm -f commandlinetools-*.zip

ENV PATH="${ANDROID_HOME}/cmdline-tools/cmdline-tools/bin:${PATH}"
RUN yes | sdkmanager --licenses \
	&& sdkmanager "platform-tools" "build-tools;34.0.0" "platforms;android-34" "cmdline-tools;latest" "cmake;3.10.2.4988404" "ndk;23.2.8568313"

# ---

FROM base AS rcedit

RUN wget https://github.com/electron/rcedit/releases/download/v2.0.0/rcedit-x64.exe -O /opt/rcedit.exe

# ---

# Final stage
FROM base

RUN gpg --homedir /tmp --no-default-keyring --keyring /usr/share/keyrings/mono-official-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 3FA7E0328081BFF6A14DA29AA6A19B38D3D831EF
RUN echo "deb [signed-by=/usr/share/keyrings/mono-official-archive-keyring.gpg] https://download.mono-project.com/repo/debian stable-buster main" > /etc/apt/sources.list.d/mono-official-stable.list

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	apt-get update && apt-get install --no-install-recommends -y \
	git \
	git-lfs \
	osslsigncode \
	rsync \
	wine64 \
	mono-devel

COPY --from=android /opt/android-sdk /opt/android-sdk
COPY --from=rcedit /opt/rcedit.exe /opt/rcedit.exe

ENV GODOT_ANDROID_KEYSTORE_DEBUG_PATH=/srv/android/debug.keystore
ENV GODOT_ANDROID_KEYSTORE_RELEASE_PATH=/srv/android/release.keystore

COPY command.sh /opt/command.sh
COPY editor_settings.tres /opt/editor_settings.tres

CMD [ "/opt/command.sh" ]
