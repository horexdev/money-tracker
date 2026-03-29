import {
  ForkKnife,
  Car,
  FilmSlate,
  ShoppingBag,
  FirstAid,
  Money,
  Laptop,
  House,
  Coffee,
  DeviceMobile,
  Lightning,
  Heartbeat,
  GraduationCap,
  Airplane,
  Gift,
  GameController,
  PawPrint,
  TShirt,
  Barbell,
  MusicNote,
  Bank,
  Wallet,
  CurrencyDollar,
  Basket,
  Bus,
  Taxi,
  Bed,
  Baby,
  Flower,
  Pizza,
  Wine,
  BookOpen,
  Briefcase,
  Camera,
  Wrench,
  Scissors,
  ShieldCheck,
  Star,
  Tag,
  type IconWeight,
} from '@phosphor-icons/react'
import type { ReactNode } from 'react'

type IconComponent = React.ComponentType<{ size?: number; weight?: IconWeight; className?: string }>

/** Emoji → Phosphor icon mapping (for legacy emoji-based categories) */
const EMOJI_MAP: Record<string, IconComponent> = {
  // Food
  '🍔': ForkKnife, '🍕': Pizza, '🍳': ForkKnife, '🥗': ForkKnife,
  '🍜': ForkKnife, '🍱': ForkKnife, '🥘': ForkKnife,
  // Transport
  '🚕': Taxi, '🚗': Car, '🚌': Bus, '🚎': Bus, '🏍️': Car, '🏍': Car,
  '🚂': Car, '⛽': Car,
  // Entertainment
  '🎬': FilmSlate, '🎭': FilmSlate, '🎪': FilmSlate, '🎮': GameController,
  '🎵': MusicNote, '🎶': MusicNote,
  // Shopping
  '🛍️': ShoppingBag, '🛍': ShoppingBag, '🛒': ShoppingBag,
  // Health
  '💊': FirstAid, '🏥': FirstAid, '🩺': FirstAid,
  // Money / income
  '💰': Money, '💵': Money, '💸': Money, '💲': CurrencyDollar,
  '💼': Briefcase, '🧾': Money,
  // Tech
  '💻': Laptop, '🖥️': Laptop, '📱': DeviceMobile, '📲': DeviceMobile,
  // Home
  '🏠': House, '🏡': House, '🛋️': Bed, '🛋': Bed,
  // Other
  '📦': Tag, '📫': Tag, '📬': Tag,
  '☕': Coffee, '⚡': Lightning, '❤️': Heartbeat, '🎓': GraduationCap,
  '✈️': Airplane, '🎁': Gift, '🐾': PawPrint, '👕': TShirt,
  '💪': Barbell, '🏦': Bank, '👛': Wallet, '🌸': Flower,
  '🌺': Flower, '🐶': PawPrint, '🐱': PawPrint, '👶': Baby,
  '📚': BookOpen, '📖': BookOpen, '🔧': Wrench, '✂️': Scissors,
  '📷': Camera, '📸': Camera, '🔒': ShieldCheck,
}

/** All available icons for the icon picker, keyed by a stable string ID */
export const ICON_CHOICES: { id: string; Icon: IconComponent; label: string }[] = [
  { id: 'fork-knife', Icon: ForkKnife, label: 'Food' },
  { id: 'coffee', Icon: Coffee, label: 'Coffee' },
  { id: 'pizza', Icon: Pizza, label: 'Pizza' },
  { id: 'wine', Icon: Wine, label: 'Drinks' },
  { id: 'basket', Icon: Basket, label: 'Groceries' },
  { id: 'car', Icon: Car, label: 'Car' },
  { id: 'bus', Icon: Bus, label: 'Bus' },
  { id: 'taxi', Icon: Taxi, label: 'Taxi' },
  { id: 'airplane', Icon: Airplane, label: 'Travel' },
  { id: 'house', Icon: House, label: 'Home' },
  { id: 'bed', Icon: Bed, label: 'Rent' },
  { id: 'shopping-bag', Icon: ShoppingBag, label: 'Shopping' },
  { id: 'film-slate', Icon: FilmSlate, label: 'Entertainment' },
  { id: 'game-controller', Icon: GameController, label: 'Games' },
  { id: 'music-note', Icon: MusicNote, label: 'Music' },
  { id: 'first-aid', Icon: FirstAid, label: 'Health' },
  { id: 'heartbeat', Icon: Heartbeat, label: 'Fitness' },
  { id: 'barbell', Icon: Barbell, label: 'Gym' },
  { id: 'graduation-cap', Icon: GraduationCap, label: 'Education' },
  { id: 'book-open', Icon: BookOpen, label: 'Books' },
  { id: 'laptop', Icon: Laptop, label: 'Tech' },
  { id: 'device-mobile', Icon: DeviceMobile, label: 'Phone' },
  { id: 'money', Icon: Money, label: 'Money' },
  { id: 'currency-dollar', Icon: CurrencyDollar, label: 'Salary' },
  { id: 'wallet', Icon: Wallet, label: 'Wallet' },
  { id: 'bank', Icon: Bank, label: 'Bank' },
  { id: 'briefcase', Icon: Briefcase, label: 'Work' },
  { id: 'gift', Icon: Gift, label: 'Gifts' },
  { id: 'paw-print', Icon: PawPrint, label: 'Pets' },
  { id: 'baby', Icon: Baby, label: 'Kids' },
  { id: 't-shirt', Icon: TShirt, label: 'Clothes' },
  { id: 'lightning', Icon: Lightning, label: 'Utilities' },
  { id: 'wrench', Icon: Wrench, label: 'Repairs' },
  { id: 'scissors', Icon: Scissors, label: 'Beauty' },
  { id: 'camera', Icon: Camera, label: 'Photo' },
  { id: 'flower', Icon: Flower, label: 'Garden' },
  { id: 'shield-check', Icon: ShieldCheck, label: 'Insurance' },
  { id: 'star', Icon: Star, label: 'Other' },
  { id: 'tag', Icon: Tag, label: 'Tags' },
]

/** Lookup map: icon ID → component */
const ICON_ID_MAP: Record<string, IconComponent> = {}
for (const choice of ICON_CHOICES) {
  ICON_ID_MAP[choice.id] = choice.Icon
}

/**
 * Render a category icon. Accepts either:
 * - A Phosphor icon ID (e.g. "fork-knife") — new format
 * - An emoji string (e.g. "🍔") — legacy format, mapped to Phosphor icon or rendered as emoji fallback
 */
export function CategoryIcon({
  emoji,
  size = 20,
  weight = 'fill',
  className = '',
}: {
  emoji: string
  size?: number
  weight?: IconWeight
  className?: string
}): ReactNode {
  // Try icon ID first (new format)
  const IconById = ICON_ID_MAP[emoji]
  if (IconById) {
    return <IconById size={size} weight={weight} className={className} />
  }
  // Try emoji mapping (legacy format)
  const IconByEmoji = EMOJI_MAP[emoji]
  if (IconByEmoji) {
    return <IconByEmoji size={size} weight={weight} className={className} />
  }
  // Fallback: render as emoji text
  return <span style={{ fontSize: size * 0.9, lineHeight: 1 }}>{emoji}</span>
}
