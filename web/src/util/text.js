const toTitleCase = (str) => str.toLowerCase().replace(/(?:^|\s)\S/g, (char) => char.toUpperCase());

/**
 * The restaurants write their dish names in whatever case they feel like
 * ("PIZZA TRICOLORE"), and quote them inconsistently. Anything that shows a dish
 * name to the user goes through here.
 */
export const dishTitle = (name) => toTitleCase(name.replace(/[«»"]/g, '').trim());
