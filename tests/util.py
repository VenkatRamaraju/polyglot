#!/usr/bin/env python3

# Imports
from functools import lru_cache
import spacy
from hebrew_tokenizer import tokenize
from indicnlp.tokenize import indic_tokenize
from pyvi import ViTokenizer
import re
from camel_tools.tokenizers.word import simple_word_tokenize
from pythainlp.tokenize import word_tokenize
import jieba
import fugashi

# Global
english_nlp = spacy.load("en_core_web_sm", disable=["parser", "ner"])
russian_nlp = spacy.load("ru_core_news_sm", disable=["parser", "ner"])
japanese_tagger = fugashi.GenericTagger()

@lru_cache(maxsize=None)
def get_segmenter(lang):
    if lang == "en":
        return lambda t: [tok.text for tok in english_nlp(t)]

    if lang == "ru":
        return lambda t: [tok.text for tok in russian_nlp(t)]

    if lang == "he":
        return lambda txt: [t[0] for t in tokenize(txt)]

    if lang == "bn":
        return lambda t: indic_tokenize.trivial_tokenize(t, lang="bn")

    if lang == "vi":
        return lambda t: ViTokenizer.tokenize(t).split()

    if lang == "ko":
        return lambda text: re.findall(r"[가-힣]+|[^\s가-힣]", text)

    if lang == "ar":
        return simple_word_tokenize

    if lang == "th":
        return lambda t: word_tokenize(t, keep_whitespace=False)

    if lang == "zh-Hans":
        return lambda t: list(jieba.cut(t))

    if lang == "ja":
        return lambda t: [tok.surface for tok in japanese_tagger(t)]

    # Fallback — whitespace split 
    return lambda t: t.split()

def count_words_batch(sentences, lang: str, return_dict=False):
    segment = get_segmenter(lang)
    if return_dict:
        return {s: len(segment(s)) for s in sentences}
    return [len(segment(s)) for s in sentences]
