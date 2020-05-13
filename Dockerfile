FROM ubuntu:20.04

RUN sed -i 's|http://archive.ubuntu.com/ubuntu/|mirror://mirrors.ubuntu.com/mirrors.txt|g' /etc/apt/sources.list \
 && apt update && apt install -y --no-install-recommends libdlib19 && apt clean \
 && adduser --system --group --no-create-home --disabled-login --uid 2000 user

EXPOSE 8002
USER user
ENTRYPOINT ["kpopnetd", "-H", "0.0.0.0", "-m", "/models", "--cfg", "/kpopnet.toml"]
COPY kpopnetd /usr/bin/
COPY testdata/models/shape_predictor_5_face_landmarks.dat \
     testdata/models/dlib_face_recognition_resnet_model_v1.dat \
     testdata/models/mmod_human_face_detector.dat \
     /models/
