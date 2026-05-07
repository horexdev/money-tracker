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
  ArrowsLeftRight,
  PiggyBank,
  Scales,
  type IconWeight,
} from '@phosphor-icons/react'
import type { ReactNode } from 'react'

type IconComponent = React.ComponentType<{ size?: number; weight?: IconWeight; className?: string }>

/** All available icons for the icon picker, keyed by a stable string ID */
// eslint-disable-next-line react-refresh/only-export-components
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
  { id: 'piggy-bank', Icon: PiggyBank, label: 'Savings' },
  { id: 'arrows-left-right', Icon: ArrowsLeftRight, label: 'Transfer' },
  { id: 'scales', Icon: Scales, label: 'Adjustment' },
]

/** Lookup map: icon ID → component */
const ICON_ID_MAP: Record<string, IconComponent> = {}
for (const choice of ICON_CHOICES) {
  ICON_ID_MAP[choice.id] = choice.Icon
}

/**
 * Render a category icon by its Phosphor icon ID (e.g. "fork-knife").
 * Falls back to Star if the ID is not recognized.
 */
export function CategoryIcon({
  icon,
  size = 20,
  weight = 'fill',
  className = '',
}: {
  icon: string
  size?: number
  weight?: IconWeight
  className?: string
}): ReactNode {
  const IconComponent = ICON_ID_MAP[icon] ?? Star
  return <IconComponent size={size} weight={weight} className={className} />
}
