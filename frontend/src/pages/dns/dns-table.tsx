import { FC } from "react";

import {
  Button,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverContent,
  PopoverTrigger,
  Table,
  Tag,
  Tbody,
  Td,
  Th,
  Thead,
  Tooltip,
  Tr,
  useColorModeValue,
} from "@chakra-ui/react";

import { DnsType } from "./dns.schema";
import { getDnsColor } from "./dns.helpers";
import { FaEye } from "react-icons/fa";
import Code from "../../common/code/code";

import type { DnsResponse } from "./dns.view";

type DnsTableProps = {
  record: DnsResponse[DnsType];
};

const DnsTableHead: FC<DnsTableProps> = ({ record }) => {
  const wideValue = !record.find((item) => typeof item.value !== "string");

  return (
    <Thead>
      <Tr>
        <Th border="none">Type</Th>

        <Th border="none">Name</Th>

        <Th border="none">TTL</Th>

        {record.find((item) => typeof item.priority !== "undefined") && (
          <Th border="none">Priority</Th>
        )}

        <Th border="none" colspan={wideValue ? 2 : 1}>
          Value
        </Th>
      </Tr>
    </Thead>
  );
};

const DnsTableRow: FC<{ row: DnsResponse[DnsType][number] }> = ({ row }) => {
  const arrowColor = useColorModeValue("gray.200", "gray.900");

  return (
    <Tr>
      <Td border="none">
        <Tag size="sm" fontWeight="bold" colorScheme={getDnsColor(row.type)}>
          {row.type}
        </Tag>
      </Td>

      <Td border="none" isTruncated>
        {row.name}
      </Td>

      <Td border="none">{row.ttl}</Td>

      {typeof row.priority !== "undefined" && (
        <Td border="none">{row.priority}</Td>
      )}

      {typeof row.value === "string" && (
        <Td border="none" colspan={2} height="32px">
          <Code isTruncated>{row.value}</Code>
        </Td>
      )}

      {typeof row.value !== "string" && (
        <Td border="none">
          <Popover>
            <PopoverTrigger>
              <Button size="sm" leftIcon={<FaEye />}>
                View value
              </Button>
            </PopoverTrigger>
            <PopoverContent
              width="auto"
              maxW="95vw"
              _focus={{ outline: "none" }}
            >
              <PopoverArrow bg={arrowColor} />
              <PopoverBody p={0} bg={arrowColor} borderRadius="md">
                <Code>{JSON.stringify(row.value, null, 2)}</Code>
              </PopoverBody>
            </PopoverContent>
          </Popover>
        </Td>
      )}
    </Tr>
  );
};

const DnsTable: FC<DnsTableProps> = ({ record }) => {
  return (
    <Table variant="simple" w="100%" size="sm" minW="580px">
      <DnsTableHead record={record} />

      <Tbody>
        {record.map((row) => (
          <DnsTableRow row={row} />
        ))}
      </Tbody>
    </Table>
  );
};

export default DnsTable;