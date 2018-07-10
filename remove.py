import numpy as np
import cv2 as cv
img = cv.imread('image.jpg')
mask = cv.imread('mask.png', 0)
dst = cv.inpaint(img, mask, 3, cv.INPAINT_TELEA)
cv.imwrite('image_r.jpg', dst)
cv.waitKey(0)
cv.destroyAllWindows()
