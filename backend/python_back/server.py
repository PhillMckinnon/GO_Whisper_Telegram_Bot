from sanic import Sanic
from sanic.response import json
from sanic_ext import Extend
from sanic.response import raw
import os
import tempfile
import whisper
from pydub import AudioSegment
from py_synth_file import synthesize_audio
from py_trans_file import transcribe_audio
from py_detectlang_file import detect_language
from dotenv import load_dotenv
load_dotenv()
app = Sanic("Backend_Speech_Server")
app.config.CORS_ORIGINS = os.getenv("CORS_ORIGIN", 'http://localhost:55756')

app.config.CORS_METHODS = ["POST", "GET", "OPTIONS"]
app.config.CORS_ALLOW_HEADERS = ["Content-Type", "Authorization"]
app.config.CORS_SUPPORTS_CREDENTIALS = True
app.config.CORS_AUTOMATIC_OPTIONS = True
app.config.CORS_ALWAYS_SEND = True
app.config.CORS_VARY_HEADER = True
Extend(app)
MAX_SIZE_MB = int(os.getenv("MAX_SIZE_MB", 10))
MAX_DURATION = int(os.getenv("MAX_DURATION", 120))
@app.post("/api/detect")
async def detect(request):
    if "file" not in request.files:
        return json({"error": "A file is required."}, status=400)
    file_obj = request.files["file"][0]
    file_bytes = file_obj.body
    filename = file_obj.name
    if len(file_bytes) > MAX_SIZE_MB * 1024 * 1024:
        return json({"error": "File too large (max 20MB)"}, status=400)
    try:
        model = whisper.load_model("small", device="cpu")
        result = detect_language(file_bytes, filename, model)
        return result
    except Exception as e:
        return json({"error": f"Audio conversion failed: {str(e)}"}, status=500)

@app.post("/api/transcribe")
async def handle_transcribe(request):
    if "file" not in request.files:
        return json({"error": "A file is required."}, status=400)
    file_obj = request.files["file"][0]
    file_bytes = file_obj.body
    filename = file_obj.name
    if len(file_bytes) > MAX_SIZE_MB * 1024 * 1024:
        return json({"error": "File too large (max 20MB)"}, status=400)
    try:
        model = whisper.load_model("small", device="cpu")
        result = transcribe_audio(file_bytes, filename, model)
        return result
    except Exception as e:
        return json({"error": f"Audio conversion failed: {str(e)}"}, status=500)

@app.post("/api/synthesize")
async def handle_synthesize(request):
    if "file" not in request.files or "text" not in request.form:
        return json({"error": "a File is required."}, status=400)
    file_obj = request.files["file"][0]
    text_input = request.form["text"][0]
    file_bytes = file_obj.body
    filename = file_obj.name
    language = request.form.get("language", [None][0])
    if len(file_bytes) > MAX_SIZE_MB * 1024 * 1024:
        return json({"error": "File exceeds the limit (20MB)"}, status=400)
    try:
        with tempfile.TemporaryDirectory() as tmpdir:
            input_path = os.path.join(tmpdir, filename)
            wav_path = os.path.join(tmpdir, "converted.wav")
            output_path = os.path.join(tmpdir, "output.wav")
            with open(input_path, "wb") as f:
                f.write(file_bytes)
            audio = AudioSegment.from_file(input_path)
            duration_sec = len(audio) / 1000
            if duration_sec > MAX_DURATION:
                return json({"error": f"Audio exceeds the {MAX_DURATION} second limit."}, status=400)
            audio.export(wav_path, format="wav")
            if not os.path.exists(wav_path):
                return json({"error": "WAV conversion failed."}, status=400)
            synthesize_audio(text_input, wav_path, output_path, language)
            with open(output_path, "rb") as f:
                output_bytes = f.read()
        return raw(
            output_bytes,
            headers={
                "Content-Disposition": "attachment; filename=synthesized_output.wav",
                "Content-Type": "audio/wav",
            }
        )
    except Exception as e:
        return json({"error": f"Audio conversion failed: {str(e)}"}, status=500)

