import endpoint from "./endpoint";

test("endpoint", () => {
  expect(endpoint("/:id", { id: 42 })).toEqual("/42/");
  expect(endpoint("/:id", { id: 42, location: "123" })).toEqual(
    "/42?location=123"
  );
  expect(endpoint("/my-endpoint", { id: 42, location: "123" })).toEqual(
    "/my-endpoint?id=42&location=123"
  );
  expect(endpoint("/:date", { date: "01/02/1994" })).toEqual("/01%2F02%2F1994");
});
