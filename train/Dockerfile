FROM python:3.9.15-slim-buster

COPY datagen.py /datagen.py

RUN pip install openai==0.9.4

ENV OPENAI_API_KEY ""

VOLUME ["/data"]

ENTRYPOINT ["openai"]
