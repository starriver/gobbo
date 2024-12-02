FROM debian:bullseye-slim

# Largely stolen from:
# https://github.com/abarichello/godot-ci/

ENV DEBIAN_FRONTEND=noninteractive

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	apt-get update && apt-get install --no-install-recommends -y \
	adb \
	ca-certificates \
	dirmngr \
	git \
	git-lfs \
	gnupg \
	openjdk-17-jdk-headless \
	osslsigncode \
	rsync \
	unzip \
	wget \
	wine64 \
	zip

RUN sudo gpg --homedir /tmp --no-default-keyring --keyring /usr/share/keyrings/mono-official-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 3FA7E0328081BFF6A14DA29AA6A19B38D3D831EF
RUN echo "deb [signed-by=/usr/share/keyrings/mono-official-archive-keyring.gpg] https://download.mono-project.com/repo/debian stable-buster main" | sudo tee /etc/apt/sources.list.d/mono-official-stable.list
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	apt-get update && apt-get install --no-install-recommends -y \
	mono-devel

RUN wget https://github.com/electron/rcedit/releases/download/v2.0.0/rcedit-x64.exe -O /opt/rcedit.exe



ENV ANDROID_HOME="/usr/lib/android-sdk"
RUN wget https://dl.google.com/android/repository/commandlinetools-linux-7583922_latest.zip \
	&& unzip commandlinetools-linux-*_latest.zip -d cmdline-tools \
	&& mv cmdline-tools $ANDROID_HOME/ \
	&& rm -f commandlinetools-linux-*_latest.zip

ENV PATH="${ANDROID_HOME}/cmdline-tools/cmdline-tools/bin:${PATH}"

RUN yes | sdkmanager --licenses \
	&& sdkmanager "platform-tools" "build-tools;34.0.0" "platforms;android-34" "cmdline-tools;latest" "cmake;3.10.2.4988404" "ndk;23.2.8568313"
