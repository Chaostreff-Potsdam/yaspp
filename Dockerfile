FROM python:3.13

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY *.py ./
RUN mkdir /workdir

ENTRYPOINT python yaspp.py -o /workdir /workdir/content.yaml
