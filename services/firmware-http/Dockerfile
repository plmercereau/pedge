FROM python:3.11-alpine as base

# Stage 1: Build stage
FROM base AS builder

COPY requirements.txt /requirements.txt
RUN python -m venv /venv && /venv/bin/pip install --no-cache-dir -r /requirements.txt

FROM base
COPY --from=builder /venv /venv
WORKDIR /app

COPY serve.py .

EXPOSE 5000
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl --fail http://localhost:5000/health || exit 1

CMD ["/venv/bin/gunicorn", "--bind", "0.0.0.0:5000", "serve:app"]