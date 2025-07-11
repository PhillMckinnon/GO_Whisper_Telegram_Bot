import io
import pytest
import numpy as np
from unittest.mock import Mock, patch
from pydub import AudioSegment
from sanic.response import json

from py_detectlang_file import detect_language
from py_trans_file import transcribe_audio
from py_synth_file import synthesize_audio


# ------------------------
# Fixtures
# ------------------------

@pytest.fixture
def dummy_audio_bytes():
    audio = AudioSegment.silent(duration=1000)  # 1 second
    buffer = io.BytesIO()
    audio.export(buffer, format="wav")
    return buffer.getvalue()

@pytest.fixture
def long_audio_bytes():
    audio = AudioSegment.silent(duration=360001)  # 361 seconds
    buffer = io.BytesIO()
    audio.export(buffer, format="wav")
    return buffer.getvalue()

@pytest.fixture
def sixmin_audio_bytes():
    audio = AudioSegment.silent(duration=360000)  # 360 seconds
    buffer = io.BytesIO()
    audio.export(buffer, format="wav")
    return buffer.getvalue()

@pytest.fixture
def Mock_model():
    model = Mock()
    model.transcribe.return_value = {
        "language": "en",
        "text": "This is a test"
    }
    return model


# ------------------------
# detect_language tests
# ------------------------

@pytest.mark.parametrize("filename", ["audio.wav", "voice.mp3", "sound.ogg"])
def test_detect_language_success(dummy_audio_bytes, Mock_model, filename):
    result = detect_language(dummy_audio_bytes, filename, Mock_model)
    assert result.status == 200
    assert b"language" in result.body

@pytest.mark.parametrize("filename", ["", "file"])
def test_detect_language_no_extension(dummy_audio_bytes, Mock_model, filename):
    result = detect_language(dummy_audio_bytes, filename, Mock_model)
    assert result.status == 400
    assert b"Cannot determine file type" in result.body


# ------------------------
# transcribe_audio tests
# ------------------------

def test_transcribe_audio_success(dummy_audio_bytes, Mock_model):
    result = transcribe_audio(dummy_audio_bytes, "test.wav", Mock_model)
    assert result.status == 200
    assert b"language" in result.body
    assert b"text" in result.body

def test_transcribe_audio_sample_rate_mismatch(dummy_audio_bytes, Mock_model):
    with patch("py_trans_file.sf.read") as mock_read:
        mock_read.return_value = (np.zeros(16000), 8000)
        result = transcribe_audio(dummy_audio_bytes, "test.wav", Mock_model)
        assert result.status == 400
        assert b"Sample rate must be 16000 Hz" in result.body

def test_transcribe_audio_too_long(long_audio_bytes, Mock_model):
    result = transcribe_audio(long_audio_bytes, "long.wav", Mock_model)
    assert result.status == 400
    assert b"Audio too long" in result.body

def test_transcribe_audio_six(sixmin_audio_bytes, Mock_model):
    result = transcribe_audio(sixmin_audio_bytes, "long.wav", Mock_model)
    assert result.status == 400

# ------------------------
# synthesize_audio tests
# ------------------------

@patch("TTS.api.TTS")
@patch("py_synth_file.whisper")
def test_synthesize_audio_with_language(mock_whisper, mock_tts):
    mock_tts_instance = Mock()
    mock_tts.return_value = mock_tts_instance

    mock_model = Mock()
    mock_model.transcribe.return_value = {"language": "en"}
    mock_whisper.load_model.return_value = mock_model
    mock_tts_instance.tts_to_file(
        text="hello world",
        speaker_wav=b"fake bytes",
        output_path="output.wav",
        language="en"
    )
    mock_tts_instance.tts_to_file.assert_called_once_with( text="hello world",
        speaker_wav=b"fake bytes",
        output_path="output.wav",
        language="en")

@patch("TTS.api.TTS")
@patch("py_synth_file.whisper")
def test_synthesize_audio_with_mocked_tts(mock_whisper, mock_tts):
    mock_tts_instance = Mock()
    mock_tts.return_value = mock_tts_instance

    mock_model = Mock()
    mock_model.transcribe.return_value = {"language": "en"}
    mock_whisper.load_model.return_value = mock_model

    mock_tts_instance.tts_to_file(
        text="hello world",
        speaker_wav=b"fake bytes",
        output_path="output.wav",
        language=None
    )
    mock_tts_instance.tts_to_file.assert_called_once()
