import type { ButtonHTMLAttributes, PropsWithChildren } from "react";

type Props = PropsWithChildren<
  ButtonHTMLAttributes<HTMLButtonElement> & {
    variant?: "primary" | "secondary";
  }
>;

export function Button({ variant = "primary", className = "", ...props }: Props) {
  const base =
    "inline-flex items-center justify-center rounded-md px-3 py-2 text-sm font-medium transition border";
  const variants =
    variant === "primary"
      ? "bg-black text-white border-black hover:bg-neutral-800"
      : "bg-white text-black border-neutral-300 hover:bg-neutral-50";
  return <button className={`${base} ${variants} ${className}`} {...props} />;
}

