FROM pytorch/pytorch:latest AS app
WORKDIR /app

RUN --mount=type=bind,source=./requirements.txt,target=/app/requirements.txt pip install -r requirements.txt
