import os
import sys
from nowatermark import WatermarkRemover

if len(sys.argv) < 2:
    print("Invalid arguments: nwm.py <image>")
    sys.exit(1)

filepath = os.path.abspath(sys.argv[1])
file_name, file_ext = os.path.splitext(filepath)


mask = os.path.abspath('./mask.png')
remover = WatermarkRemover()
remover.load_watermark_template(mask)

remover.remove_watermark(filepath, file_name + '_result' + file_ext)
