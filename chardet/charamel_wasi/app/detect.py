import sys

from charamel import Detector, Encoding


ENCODINGS = (
    Encoding.UTF_8,
    Encoding.UTF_16,
    Encoding.UTF_16_BE,
    Encoding.UTF_16_LE,
    Encoding.UTF_32,
    Encoding.UTF_32_BE,
    Encoding.UTF_32_LE,
    Encoding.ASCII,
    Encoding.LATIN_1,
    Encoding.ISO_8859_2,
    Encoding.ISO_8859_3,
    Encoding.ISO_8859_4,
    Encoding.ISO_8859_5,
    Encoding.ISO_8859_6,
    Encoding.ISO_8859_7,
    Encoding.ISO_8859_8,
    Encoding.ISO_8859_9,
    Encoding.ISO_8859_10,
    Encoding.ISO_8859_11,
    Encoding.ISO_8859_13,
    Encoding.ISO_8859_14,
    Encoding.ISO_8859_15,
    Encoding.ISO_8859_16,
    Encoding.CP_1250,
    Encoding.CP_1251,
    Encoding.CP_1252,
    Encoding.CP_1253,
    Encoding.CP_1254,
    Encoding.CP_1255,
    Encoding.CP_1256,
    Encoding.CP_1257,
    Encoding.CP_1258,
    Encoding.GB_2312,
    Encoding.GB_K,
    Encoding.GB_18030,
    Encoding.BIG_5,
    Encoding.BIG_5_HKSCS,
    Encoding.HZ,
    Encoding.EUC_JP,
    Encoding.SHIFT_JIS,
    Encoding.ISO_2022_JP,
    Encoding.EUC_KR,
    Encoding.CP_949,
    Encoding.ISO_2022_KR,
    Encoding.KOI_8_R,
    Encoding.KOI_8_U,
    Encoding.CP_866,
    Encoding.TIS_620,
    Encoding.CP_874,
    Encoding.MAC_ROMAN,
    Encoding.MAC_CYRILLIC,
)


data = sys.stdin.buffer.read()
detector = Detector(ENCODINGS)
encoding = detector.detect(data)
sys.stdout.write(encoding.value if encoding is not None else "")
