export function rgbHex(channels: number[]) {
  return (
    "#" +
    channels
      .map((value) =>
        Math.min(255, Math.max(0, Math.round(value)))
          .toString(16)
          .padStart(2, "0"),
      )
      .join("")
  );
}

export function hexChannels(value: string): number[] | null {
  const short = /^#([a-f\d])([a-f\d])([a-f\d])$/i.exec(value.trim());
  const hex = short
    ? "#" +
      short
        .slice(1)
        .map((c) => c + c)
        .join("")
    : value.trim();
  if (!/^#[a-f\d]{6}$/i.test(hex)) return null;
  return [1, 3, 5].map((offset) => parseInt(hex.slice(offset, offset + 2), 16));
}

// Canvas converts supported CSS colors (including OKLCH) into the picker's sRGB channels.
// The original CSS value is preserved until the user actually chooses another color.
export function cssColorHex(value: string): string | null {
  const channels = hexChannels(value);
  if (channels) return rgbHex(channels);
  if (
    typeof CSS === "undefined" ||
    !CSS.supports("color", value) ||
    /var\(|currentcolor|inherit|initial|unset/i.test(value)
  )
    return null;
  const canvas = document.createElement("canvas");
  canvas.width = canvas.height = 1;
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context) return null;
  context.fillStyle = value;
  context.fillRect(0, 0, 1, 1);
  const pixel = context.getImageData(0, 0, 1, 1).data;
  // These controls edit opaque colors; do not silently discard transparency.
  return pixel[3] === 255 ? rgbHex(Array.from(pixel.slice(0, 3))) : null;
}

export function hslHex(hue: number, saturation: number, lightness: number) {
  const s = saturation / 100,
    l = lightness / 100;
  const a = s * Math.min(l, 1 - l);
  return rgbHex(
    [0, 8, 4].map((n) => {
      const k = (n + hue / 30) % 12;
      return 255 * (l - a * Math.max(-1, Math.min(k - 3, 9 - k, 1)));
    }),
  );
}

export const colorPalette = [
  [100, 96, 90, 80, 65, 50, 35, 20, 10, 0].map((l) => hslHex(0, 0, l)),
  ...[0, 25, 45, 65, 100, 145, 175, 200, 220, 250, 280, 320].map((h) =>
    [96, 88, 76, 65, 55, 45, 36, 27, 19, 12].map((l) => hslHex(h, 78, l)),
  ),
];
