import { FC, Fragment, useMemo } from "react";

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

type DnsTableHeadProps = DnsTableProps & {
  showPriority: boolean;
};

const DnsTableHead: FC<DnsTableHeadProps> = ({ record, showPriority }) => {
  const headBg = useColorModeValue("gray.100", "gray.600");

  return (
    <Thead>
      <Tr>
        <Th p={3} pl={5} border="none" bg={headBg} borderLeftRadius="md">
          Type
        </Th>

        <Th p={3} border="none" bg={headBg} colSpan={2}>
          Name
        </Th>

        <Th p={3} border="none" bg={headBg}>
          TTL
        </Th>

        {showPriority && (
          <Th p={3} border="none" bg={headBg}>
            Priority
          </Th>
        )}

        <Th
          p={3}
          pr={5}
          border="none"
          borderRightRadius="md"
          bg={headBg}
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
  row: DnsResponse[DnsType][number]["records"][number];
  showPriority: boolean;
};

const DnsTableRow: FC<DnsTableRowProps> = ({ row, showPriority }) => {
  const arrowColor = useColorModeValue("gray.200", "gray.900");

  return (
    <Tr>
      <Td border="none">
        <Tag size="sm" fontWeight="bold" colorScheme={getDnsColor(row.type)}>
          {row.type}
        </Tag>
      </Td>

      <Td border="none" isTruncated colSpan={2}>
        {row.name}
      </Td>

      <Td border="none">{row.ttl}</Td>

      {showPriority && <Td border="none">{row.priority || ""}</Td>}

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
  const labelBg = useColorModeValue("gray.50", "gray.600");

  const showPriority = useMemo(
    () =>
      !!record.find(
        (item) =>
          !!item.records.find((rec) => typeof rec.priority !== "undefined")
      ),
    [record]
  );

  return (
    <Table variant="simple" w="100%" size="sm" minW="580px">
      <DnsTableHead record={record} showPriority={showPriority} />

      <Tbody>
        {record.map(({ server, records }) => (
          <Fragment key={server}>
            <Tr>
              <Td colSpan={showPriority ? 7 : 6} py={1} px={0} border="none">
                <Box
                  py={3}
                  px={5}
                  w="100%"
                  bg={labelBg}
                  opacity={0.8}
                  borderRadius="md"
                >
                  <Text fontFamily="monospace" fontWeight="bold">
                    {server}
                  </Text>
                </Box>
              </Td>
            </Tr>

            {records.map((row, index) => (
              <DnsTableRow key={index} row={row} showPriority={showPriority} />
            ))}
          </Fragment>
        ))}
      </Tbody>
    </Table>
  );
};

export default DnsTable;
