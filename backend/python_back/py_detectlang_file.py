import whisper
import numpy as np
import io
import os
import soundfile as sf
from pydub import AudioSegment
from sanic.response import json
MAX_SIZE_MB = int(os.getenv("MAX_SIZE_MB", 10))
MAX_DURATION = int(os.getenv("MAX_DURATION", 120))  # 6 minutes
def detect_language(file_bytes: bytes, filename: str, model):
    ext = os.path.splitext(filename)[-1].lower().replace(".", "")
    if not ext:
        return json({"error": "Cannot determine file type"}, status=400)
    try:
        audio = AudioSegment.from_file(io.BytesIO(file_bytes))
    except Exception as e:
        return json({"error": f"Audio decoding failed: {str(e)}"}, status=400)
    wav_io = io.BytesIO()
    audio.set_frame_rate(16000).set_channels(1).export(wav_io, format="wav")
    wav_io.seek(0)
    audio_np, sample_rate = sf.read(wav_io)
    if sample_rate != whisper.audio.SAMPLE_RATE:
        return json({"error": f"Sample rate must be 16000 Hz, got {sample_rate}"}, status=400)
    duration = len(audio_np) / sample_rate
    if duration > MAX_DURATION:
        return json({"error": f"Audio too long: {duration:.1f}s. Max allowed is {MAX_DURATION}s."}, status=400)
    if audio_np.ndim > 1:
        audio_np = np.mean(audio_np, axis=1)
    audio_np = audio_np.astype(np.float32)
    result = model.transcribe(audio_np)
    return json({"language": result.get("language")})