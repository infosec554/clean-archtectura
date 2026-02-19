#!/bin/bash

# === see_changes.sh ===
# Foydalanuvchining so‘nggi commitlarini va o‘zgartirgan fayllarini ko‘rsatadi
# Ishlatish: ./see_changes.sh Oybek

AUTHOR=$1

if [ -z "$AUTHOR" ]; then
  echo "⚠️  Foydalanuvchi ismini kiriting. Masalan:"
  echo "./see_changes.sh Oybek"
  exit 1
fi

echo "🔍 $AUTHOR tomonidan so‘nggi commitlar: "
echo "---------------------------------------"
git log --author="$AUTHOR" --pretty=format:"%C(yellow)%h %Cgreen%ad %Creset%s" --date=short -n 10
echo ""

LAST_COMMIT=$(git log --author="$AUTHOR" --pretty=format:"%h" -n 1)

if [ -z "$LAST_COMMIT" ]; then
  echo "❌ $AUTHOR tomonidan commit topilmadi."
  exit 0
fi

echo "📂 So‘nggi commitdagi fayllar:"
git show --name-only --oneline $LAST_COMMIT | tail -n +2
echo ""
echo "✅ Tugadi."
