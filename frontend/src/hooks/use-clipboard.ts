import { useToast } from "@chakra-ui/react";

export const useClipboard = (textToCopy: string) => {
  const toast = useToast();

  const copyToClipboard = () => {
    navigator.clipboard.writeText(textToCopy);
    toast({
      status: "success",
      title: "Copied to your clipboard succesfully",
    });
  };

  return copyToClipboard;
};
