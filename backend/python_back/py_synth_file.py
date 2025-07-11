import torch
import whisper
tts_model = None
def init_tts():
    global tts_model
    if tts_model is None:
        from TTS.api import TTS

        tts_model = TTS("tts_models/multilingual/multi-dataset/xtts_v2").to("cpu")
def synthesize_audio(text, speaker_wav, output_path, language=None):
    init_tts()
    if language== None:
        from TTS.api import TTS
        whisper_model = whisper.load_model("small", device="cpu")
        result = whisper_model.transcribe(speaker_wav)
        language = result["language"]
    tts_model.tts_to_file(
        text=text,
        speaker_wav=speaker_wav,
        language=language,
        file_path=output_path
    )
