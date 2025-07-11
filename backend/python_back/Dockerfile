FROM python:3.10-slim

WORKDIR /app

COPY requirements.txt .

# Install required packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    git \
    gcc \
    g++ \
    libffi-dev \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create model directories
RUN mkdir -p /root/.local/share/tts/tts_models--multilingual--multi-dataset--xtts_v2 \
    /root/.cache/whisper

# Download Coqui XTTS v2 model
RUN curl -L -o /root/.local/share/tts/tts_models--multilingual--multi-dataset--xtts_v2/model.pth \
    https://huggingface.co/coqui/XTTS-v2/resolve/main/model.pth

# Download Whisper small model
RUN curl -L -o /root/.cache/whisper/small.pt \
    https://huggingface.co/openai/whisper-small/resolve/main/small.pt

# Install Python dependencies
RUN pip install --no-cache-dir numpy && \
    pip install --no-cache-dir -r requirements.txt && \
    rm -rf /root/.cache/pip

# Remove curl (optional cleanup)
RUN apt-get remove -y curl || true && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

# Copy application files
COPY app.py .
COPY server.py .
COPY py_detectlang_file.py .
COPY py_synth_file.py .
COPY py_trans_file.py .
COPY test_main_app_tests.py .

# Agree to Coqui TTS terms
ENV COQUI_TOS_AGREED=1

# Expose port from ENV (set via docker-compose)
EXPOSE ${PORT}

# Start backend
CMD ["python3", "app.py"]
