import { Card as ChakraCard, type CardRootProps } from "@chakra-ui/react";

export interface CardComponentProps extends CardRootProps {
	children: React.ReactNode;
}

export const Card = (props: CardComponentProps) => {
	const { children, ...rest } = props;

	return <ChakraCard.Root {...rest}>{children}</ChakraCard.Root>;
};

Card.Header = ChakraCard.Header;
Card.Body = ChakraCard.Body;
Card.Footer = ChakraCard.Footer;
Card.Title = ChakraCard.Title;
Card.Description = ChakraCard.Description;
