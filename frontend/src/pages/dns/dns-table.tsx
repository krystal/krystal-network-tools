import { FC, useMemo } from "react";

import {
  Button,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverContent,
  PopoverTrigger,
  Table,
  Tag,
  Box,
  Text,
  Tbody,
  Td,
  Th,
  Thead,
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

const showPriority = (record: DnsResponse[DnsType]) =>
  !!record.find((item) => typeof item.priority !== "undefined");

const DnsTableHead: FC<DnsTableProps> = ({ record }) => {
  return (
    <Thead>
      <Tr>
        <Th p={3} pl={5} border="none" bg="gray.100" borderLeftRadius="md">
          Type
        </Th>

        <Th p={3} border="none" bg="gray.100">
          Name
        </Th>

        <Th p={3} border="none" bg="gray.100">
          TTL
        </Th>

        {showPriority(record) && (
          <Th p={3} border="none" bg="gray.100">
            Priority
          </Th>
        )}

        <Th
          p={3}
          pr={5}
          border="none"
          borderRightRadius="md"
          bg="gray.100"
          colSpan={2}
          textAlign="right"
        >
          Value
        </Th>
      </Tr>
    </Thead>
  );
};

type DnsTableRowProps = {
  row: DnsResponse[DnsType][number];
};

const DnsTableRow: FC<DnsTableRowProps> = ({ row }) => {
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
        <Td border="none" colSpan={2} height="32px" textAlign="right">
          <Code isTruncated>{row.value}</Code>
        </Td>
      )}

      {typeof row.value !== "string" && (
        <Td border="none" colSpan={2} textAlign="right">
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
  const groupedRecords = useMemo(() => {
    return record.reduce((groups, row) => {
      const previous = groups[row.dnsServer] || [];

      return {
        ...groups,
        [row.dnsServer]: [...previous, row],
      };
    }, {} as { [k: string]: DnsResponse[DnsType] });
  }, [record]);

  return (
    <Table variant="simple" w="100%" size="sm" minW="580px">
      <DnsTableHead record={record} />

      <Tbody>
        {Object.keys(groupedRecords).map((dnsServer) => (
          <>
            <Tr>
              <Td
                colspan={showPriority(record) ? 6 : 5}
                py={1}
                px={0}
                border="none"
              >
                <Box py={3} px={5} w="100%" bg="gray.50" borderRadius="md">
                  <Text fontFamily="monospace" fontWeight="bold">
                    {dnsServer}
                  </Text>
                </Box>
              </Td>
            </Tr>

            {groupedRecords[dnsServer].map((row, index) => (
              <DnsTableRow key={index} row={row} />
            ))}
          </>
        ))}
      </Tbody>
    </Table>
  );
};

export default DnsTable;
