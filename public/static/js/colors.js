/**
 * Returns the specified number of colors evenly spread on the color spectrum
 *
 * @param   Int   n     The number of colors desired
 * @return  Array       An array containing the colors in rgb form.
 */
var generateColors = (function () {
    "use strict";

    // http://mjijackson.com/2008/02/rgb-to-hsl-and-rgb-to-hsv-color-model-conversion-algorithms-in-javascript
    /**
     * Converts an RGB color value to HSL. Conversion formula
     * adapted from http://en.wikipedia.org/wiki/HSL_color_space.
     * Assumes r, g, and b are contained in the set [0, 255] and
     * returns h, s, and l in the set [0, 255].
     *
     * @param   Array   rgb     The red, green, blue color values array
     * @return  Array           The HSL representation
     */
    var rgbToHsl = function (rgb) {
        var r = rgb[0] / 255,
            g = rgb[1] / 255,
            b = rgb[2] / 255,
            max = Math.max(r, g, b),
            min = Math.min(r, g, b),
            d = max - min,
            h,
            s,
            l = (max + min) / 2;

        if (max === min) {
            h = s = 0; // achromatic
        } else {
            s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
            switch (max) {
            case r:
                h = (g - b) / d + (g < b ? 6 : 0);
                break;
            case g:
                h = (b - r) / d + 2;
                break;
            case b:
                h = (r - g) / d + 4;
                break;
            }
            h /= 6;
        }

        return [Math.round(h * 255), Math.round(s * 255), Math.round(l * 255)];
    };

    /**
     * Converts an HSL color value to RGB. Conversion formula
     * adapted from http://en.wikipedia.org/wiki/HSL_color_space.
     * Assumes h, s, and l are contained in the set [0, 255] and
     * returns r, g, and b in the set [0, 255].
     *
     * @param   Array   hsl     The hue, saturation, lightness values array
     * @return  Array           The RGB representation
     */
    var hslToRgb = function (hsl) {
        var h = hsl[0] / 255,
            s = hsl[1] / 255,
            l = hsl[2] / 255,
            q = l < 0.5 ? l * (1 + s) : l + s - l * s,
            p = 2 * l - q,
            r,
            g,
            b,
            hue2rgb = function (p, q, t) {
                if (t < 0) { t += 1; }
                if (t > 1) { t -= 1; }
                if (t < 1 / 6) { return p + (q - p) * 6 * t; }
                if (t < 1 / 2) { return q; }
                if (t < 2 / 3) { return p + (q - p) * (2 / 3 - t) * 6; }
                return p;
            };

        if (s === 0) {
            r = g = b = l; // achromatic
        } else {
            r = hue2rgb(p, q, h + 1 / 3);
            g = hue2rgb(p, q, h);
            b = hue2rgb(p, q, h - 1 / 3);
        }

        return [Math.round(r * 255), Math.round(g * 255), Math.round(b * 255)];
    };

    return function (n) {
        var baseColor = rgbToHsl([138, 38, 226]),
            baseHue = baseColor[0],
            step = (240.0 / n),
            i,
            nextColor,
            colors = [];

        colors.push(hslToRgb(baseColor));

        for (i = 1; i < n; i += 1) {
            nextColor = baseColor;
            nextColor[0] = (baseHue + step * i) % 240.0;
            colors.push(hslToRgb(nextColor));
        }

        return colors;
    };
}());
