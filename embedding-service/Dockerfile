FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
# Create an empty requirements file if it doesn't exist to prevent failure
RUN touch requirements.txt && pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000
# Placeholder run command until Fastapi is set up
CMD ["tail", "-f", "/dev/null"]
