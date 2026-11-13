#!/usr/bin/env python3
"""Watermark discret des photos produits Marexstore.

- Cible : les images référencées dans data/store.db (table products, colonne image).
- Bandeau "store.neva-ci.pro" en bas à droite, blanc semi-transparent + ombre.
- Métadonnées EXIF supprimées (hygiène + anti-fuites GPS).
- Formats : JPG/PNG/WebP (Pillow). AVIF ignoré (support Pillow absent) et signalé.
- Usage échantillon : ./tools/watermark.py --sample 6
- Usage réel      : ./tools/watermark.py --apply [--db data/store.db --images images]
"""
import argparse
import os
import sqlite3
import sys

from PIL import Image, ImageDraw, ImageFont

TEXT = "store.neva-ci.pro"
FONT_PATHS = [
    "/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
    "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
]


def load_font(size):
    for p in FONT_PATHS:
        if os.path.exists(p):
            try:
                return ImageFont.truetype(p, size)
            except Exception:
                pass
    return ImageFont.load_default()


def watermark(im):
    base = im.convert("RGBA")
    w, h = base.size
    size = max(14, w // 28)
    font = load_font(size)
    layer = Image.new("RGBA", base.size, (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    bbox = d.textbbox((0, 0), TEXT, font=font)
    tw, th = bbox[2] - bbox[0], bbox[3] - bbox[1]
    pad = max(8, w // 100)
    x, y = w - tw - pad * 2, h - th - pad * 2
    # pastille sombre + texte blanc semi-transparent
    d.rectangle([x - pad, y - pad, x + tw + pad, y + th + pad], fill=(0, 0, 0, 90))
    d.text((x, y), TEXT, font=font, fill=(255, 255, 255, 200))
    out = Image.alpha_composite(base, layer)
    return out


def save(im, path):
    ext = os.path.splitext(path)[1].lower()
    if ext in (".jpg", ".jpeg"):
        im.convert("RGB").save(path, "JPEG", quality=90)
    elif ext == ".png":
        im.save(path, "PNG")
    elif ext == ".webp":
        im.save(path, "WEBP", quality=90)
    else:
        raise ValueError("format non supporté : " + ext)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--db", default="data/store.db")
    ap.add_argument("--images", default="images")
    ap.add_argument("--sample", type=int, default=0,
                    help="N images traitées vers --outdir (aperçu, sans toucher aux originaux)")
    ap.add_argument("--outdir", default="/tmp/opencode/wm-sample")
    ap.add_argument("--apply", action="store_true",
                    help="applique le watermark sur les originaux (irréversible sans backup)")
    args = ap.parse_args()

    c = sqlite3.connect(args.db)
    imgs = [r[0] for r in c.execute("SELECT DISTINCT image FROM products WHERE image<>''")]
    c.close()

    skipped, done = [], 0
    targets = imgs[: args.sample] if args.sample else imgs
    for rel in targets:
        p = os.path.join(args.images, os.path.basename(rel)) if not rel.startswith("images/") else rel
        if not os.path.exists(p):
            skipped.append((rel, "fichier absent"))
            continue
        ext = os.path.splitext(p)[1].lower()
        if ext == ".avif":
            skipped.append((rel, "avif non supporté par Pillow"))
            continue
        try:
            im = Image.open(p)
            out = watermark(im)
            if args.sample:
                os.makedirs(args.outdir, exist_ok=True)
                save(out, os.path.join(args.outdir, os.path.basename(p)))
            else:
                if not args.apply:
                    continue
                save(out, p)
            done += 1
        except Exception as e:  # noqa: BLE001
            skipped.append((rel, str(e)))
    print(f"traitées : {done} | ignorées : {len(skipped)}")
    for rel, why in skipped[:20]:
        print(f"  - {rel} : {why}")
    if not args.sample and not args.apply:
        print("(mode lecture seule : ajoutez --apply pour écrire, --sample N pour un aperçu)")
        return 1


if __name__ == "__main__":
    sys.exit(main() or 0)
